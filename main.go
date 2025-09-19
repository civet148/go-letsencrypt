package main

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"github.com/civet148/log"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/acme"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	SshScheme   = "ssh://"
	Version     = "0.1.0"
	ProgramName = "go-letsencrypt"
)

var (
	BuildTime = "2025-09-19"
	GitCommit = "<N/A>"
)

const (
	CmdFlag_Domain  = "domain"
	CmdFlag_Email   = "email"
	CmdFlag_Staging = "staging"
	CmdFlag_CertDir = "cert-dir"
	CmdFlag_Port    = "port"
)

const (
	// Let's Encrypt ACME v2 生产环境URL
	LetsEncryptURL = "https://acme-v02.api.letsencrypt.org/directory"
	// Let's Encrypt ACME v2 测试环境URL
	LetsEncryptStagingURL = "https://acme-staging-v02.api.letsencrypt.org/directory"
	// 默认证书保存目录
	DefaultCertDir = "./certs"
)

type CertManager struct {
	client     *acme.Client
	accountKey crypto.Signer
	certDir    string
	email      string
	staging    bool
}

func main() {

	app := &cli.App{
		Name:    ProgramName,
		Usage:   "[options] -m your@mail.com -d domain",
		Version: fmt.Sprintf("v%s %s commit %s", Version, BuildTime, GitCommit),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     CmdFlag_Domain,
				Usage:    "domain",
				Aliases:  []string{"d"},
				Required: true,
			},
			&cli.StringFlag{
				Name:     CmdFlag_Email,
				Usage:    "Let's Encrypt registry email",
				Aliases:  []string{"m"},
				Required: true,
			},
			&cli.StringFlag{
				Name:    CmdFlag_CertDir,
				Usage:   "Let's Encrypt certs directory",
				Aliases: []string{"c"},
				Value:   DefaultCertDir,
			},
			&cli.BoolFlag{
				Name:    CmdFlag_Staging,
				Usage:   "Let's Encrypt test environment",
				Aliases: []string{"s"},
			},
			&cli.StringFlag{
				Name:    CmdFlag_Port,
				Usage:   "Let's Encrypt http check port",
				Aliases: []string{"p"},
			},
		},
		Action: func(ctx *cli.Context) error {
			var domain = ctx.String(CmdFlag_Domain)
			var email = ctx.String(CmdFlag_Email)
			var certDir = ctx.String(CmdFlag_CertDir)
			var staging = ctx.Bool(CmdFlag_Staging)
			var port = ctx.String(CmdFlag_Port)

			// 创建证书管理器
			cm, err := NewCertManager(email, certDir, staging)
			if err != nil {
				return log.Errorf("创建证书管理器失败: %v", err.Error())
			}

			log.Infof("开始为域名 %s 申请Let's Encrypt证书(%v)...", domain, func() string {
				if staging {
					return "测试环境"
				}
				return "生产环境"
			}())

			// 申请证书
			err = cm.ObtainCertificate(domain, port)
			if err != nil {
				return log.Fatalf("申请证书失败: %v", err)
			}

			log.Infof("证书申请成功! 证书已保存到: %s", certDir)
			return nil
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Errorf("exit in error %s", err)
		os.Exit(1)
		return
	}
}

// NewCertManager 创建新的证书管理器
func NewCertManager(email, certDir string, staging bool) (*CertManager, error) {
	// 确保证书目录存在
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return nil, log.Errorf("创建证书目录失败: %v", err)
	}

	// 生成或加载账户私钥
	accountKey, err := loadOrGenerateAccountKey(filepath.Join(certDir, "account.key"))
	if err != nil {
		return nil, log.Errorf("处理账户私钥失败: %v", err)
	}

	// 选择ACME服务器URL
	directoryURL := LetsEncryptURL
	if staging {
		directoryURL = LetsEncryptStagingURL
	}

	// 创建ACME客户端
	client := &acme.Client{
		Key:          accountKey,
		DirectoryURL: directoryURL,
	}

	cm := &CertManager{
		client:     client,
		accountKey: accountKey,
		certDir:    certDir,
		email:      email,
		staging:    staging,
	}

	return cm, nil
}

// loadOrGenerateAccountKey 加载或生成账户私钥
func loadOrGenerateAccountKey(keyPath string) (crypto.Signer, error) {
	// 尝试加载现有私钥
	if keyData, err := os.ReadFile(keyPath); err == nil {
		block, _ := pem.Decode(keyData)
		if block != nil {
			key, err := x509.ParseECPrivateKey(block.Bytes)
			if err == nil {
				log.Infof("加载现有账户私钥")
				return key, nil
			}
		}
	}

	// 生成新的私钥
	log.Infof("生成新的账户私钥")
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	// 保存私钥到文件
	keyBytes, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, err
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: keyBytes,
	})

	if err := os.WriteFile(keyPath, keyPEM, 0600); err != nil {
		return nil, err
	}

	return key, nil
}

// ObtainCertificate 申请证书
func (cm *CertManager) ObtainCertificate(domain, port string) error {
	ctx := context.Background()

	// 注册账户
	err := cm.registerAccount(ctx)
	if err != nil {
		return log.Errorf("注册账户失败: %v", err)
	}

	// 创建订单
	order, err := cm.client.AuthorizeOrder(ctx, []acme.AuthzID{
		{Type: "dns", Value: domain},
	})
	if err != nil {
		return log.Errorf("创建订单失败: %v", err)
	}

	// 处理授权质询
	for _, authzURL := range order.AuthzURLs {
		err = cm.handleAuthorization(ctx, authzURL, port)
		if err != nil {
			return log.Errorf("处理授权质询失败: %v", err)
		}
	}

	// 生成证书私钥
	certKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return log.Errorf("生成证书私钥失败: %v", err)
	}

	// 创建证书签名请求
	csr, err := cm.createCSR(domain, certKey)
	if err != nil {
		return log.Errorf("创建CSR失败: %v", err)
	}

	// 完成订单并获取证书
	certs, _, err := cm.client.CreateOrderCert(ctx, order.FinalizeURL, csr, true)
	if err != nil {
		return log.Errorf("获取证书失败: %v", err)
	}

	// 保存证书和私钥
	err = cm.saveCertificates(domain, certs, certKey)
	if err != nil {
		return log.Errorf("保存证书失败: %v", err)
	}

	return nil
}

// registerAccount 注册ACME账户
func (cm *CertManager) registerAccount(ctx context.Context) error {
	account := &acme.Account{
		Contact: []string{"mailto:" + cm.email},
	}

	_, err := cm.client.Register(ctx, account, acme.AcceptTOS)
	if err != nil {
		// 如果账户已存在，忽略错误
		if acmeErr, ok := err.(*acme.Error); ok && acmeErr.StatusCode == 409 {
			log.Infof("账户已存在")
			return nil
		}
		return err
	}

	log.Infof("账户注册成功")
	return nil
}

// handleAuthorization 处理域名授权质询
func (cm *CertManager) handleAuthorization(ctx context.Context, authzURL, port string) error {
	// 获取授权信息
	authz, err := cm.client.GetAuthorization(ctx, authzURL)
	if err != nil {
		return err
	}

	if authz.Status == acme.StatusValid {
		log.Infof("域名 %s 已通过验证", authz.Identifier.Value)
		return nil
	}

	// 寻找HTTP-01质询
	var httpChallenge *acme.Challenge
	for _, challenge := range authz.Challenges {
		if challenge.Type == "http-01" {
			httpChallenge = challenge
			break
		}
	}

	if httpChallenge == nil {
		return log.Errorf("未找到HTTP-01质询")
	}

	// 计算质询响应
	response, err := cm.client.HTTP01ChallengeResponse(httpChallenge.Token)
	if err != nil {
		return err
	}

	// 启动HTTP服务器响应质询
	server := &http.Server{
		Addr: ":" + port,
	}

	http.HandleFunc("/.well-known/acme-challenge/"+httpChallenge.Token, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(response))
	})

	go func() {
		log.Infof("启动HTTP服务器在端口 %s 用于质询验证...", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Errorf("HTTP服务器错误: %v", err)
		}
	}()

	// 等待服务器启动
	time.Sleep(2 * time.Second)

	// 接受质询
	log.Infof("开始验证域名 %s...", authz.Identifier.Value)
	_, err = cm.client.Accept(ctx, httpChallenge)
	if err != nil {
		server.Close()
		return err
	}

	// 等待质询完成
	for i := 0; i < 30; i++ {
		authz, err = cm.client.GetAuthorization(ctx, authzURL)
		if err != nil {
			server.Close()
			return err
		}

		if authz.Status == acme.StatusValid {
			log.Infof("域名 %s 验证成功", authz.Identifier.Value)
			server.Close()
			return nil
		}

		if authz.Status == acme.StatusInvalid {
			server.Close()
			return log.Errorf("域名验证失败")
		}

		time.Sleep(2 * time.Second)
	}

	server.Close()
	return log.Errorf("域名验证超时")
}

// createCSR 创建证书签名请求
func (cm *CertManager) createCSR(domain string, certKey *ecdsa.PrivateKey) ([]byte, error) {
	template := &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName: domain,
		},
		DNSNames: []string{domain},
	}

	return x509.CreateCertificateRequest(rand.Reader, template, certKey)
}

// saveCertificates 保存证书和私钥到文件
func (cm *CertManager) saveCertificates(domain string, certs [][]byte, certKey *ecdsa.PrivateKey) error {
	// 保存证书链
	certPath := filepath.Join(cm.certDir, domain+".crt")
	certFile, err := os.Create(certPath)
	if err != nil {
		return err
	}
	defer certFile.Close()

	for _, cert := range certs {
		pem.Encode(certFile, &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert,
		})
	}

	// 保存私钥
	keyPath := filepath.Join(cm.certDir, domain+".key")
	keyBytes, err := x509.MarshalECPrivateKey(certKey)
	if err != nil {
		return err
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: keyBytes,
	})

	if err := os.WriteFile(keyPath, keyPEM, 0600); err != nil {
		return err
	}

	log.Infof("证书已保存到: %s", certPath)
	log.Infof("私钥已保存到: %s", keyPath)

	// 显示证书信息
	err = cm.displayCertInfo(certPath)
	if err != nil {
		log.Printf("显示证书信息失败: %v", err)
	}

	return nil
}

// displayCertInfo 显示证书信息
func (cm *CertManager) displayCertInfo(certPath string) error {
	certData, err := os.ReadFile(certPath)
	if err != nil {
		return err
	}

	block, _ := pem.Decode(certData)
	if block == nil {
		return log.Errorf("无法解析证书")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return err
	}

	log.Infof("=== 证书信息 ===")
	log.Infof("主题: %s", cert.Subject.CommonName)
	log.Infof("DNS名称: %s", strings.Join(cert.DNSNames, ", "))
	log.Infof("颁发者: %s", cert.Issuer.CommonName)
	log.Infof("有效期: %s 至 %s", cert.NotBefore.Format("2006-01-02 15:04:05"), cert.NotAfter.Format("2006-01-02 15:04:05"))
	log.Infof("序列号: %s", cert.SerialNumber)

	return nil
}
