package cmdRun

import (
	"bytes"
	"errors"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"time"
)

//用来执行 linux 命令
//需要完善

// SshClient 采用ssh密钥 登录服务器
type SshClient struct {
	host string

	client *ssh.Client
}

// NewSSH ssh连接服务器 root:port
func NewSSH(rootPort string) (*SshClient, error) {
	var c SshClient

	//从配置文件获取 私钥信息
	key, err := getCurUserKey()
	if err != nil {
		return nil, errors.New("从配置文件读取秘钥出错:" + err.Error())
	}

	c.host = rootPort

	//构建 ssh 登录验证方式 密钥校验
	auth := make([]ssh.AuthMethod, 0)
	signer, _ := ssh.ParsePrivateKey(key.privateKey)
	auth = append(auth, ssh.PublicKeys(signer))

	//设置连接配置信息
	cConfig := &ssh.ClientConfig{
		User:    "root", //ssh登录用户名
		Auth:    auth,   //ssh验证方式 公钥私钥 / 密码验证
		Timeout: 10 * time.Second,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	client, err := ssh.Dial("tcp", c.host, cConfig)
	if err != nil {
		return nil, errors.New("连接到服务器出错：" + err.Error())
	}

	c.client = client

	return &c, nil
}

// Run 执行
func (r *SshClient) Run(command string) (string, error) {

	s, err := r.client.NewSession()
	if err != nil {
		return "", errors.New("生成客户端会话出错：" + err.Error())
	}

	defer s.Close()

	var stdOut, stdErr bytes.Buffer

	s.Stdout = &stdOut
	s.Stderr = &stdErr

	err = s.Run(command)
	if err != nil {
		return "", err
	}

	if stdErr.String() != "" {
		return stdErr.String(), nil
	}

	return stdOut.String(), nil
}

func (r *SshClient) SftpSend(localFile, remoteFile string) error {
	cc, err := sftp.NewClient(r.client)
	if err != nil {
		return err
	}

	srcFile, err := os.Open(localFile)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := cc.Create(remoteFile)
	if err != nil {
		return err
	}
	defer destFile.Close()

	f, _ := srcFile.Stat()
	f.Size()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}

	/*//缓冲区大一点传输会快很多
	buf := make([]byte, 1024*1024)
	for {
		n, _ := srcFile.Read(buf)
		if n == 0 {
			break
		}
		_, _ = destFile.Write(buf)
	}*/

	return nil
}

// 当前用户的公钥和私钥信息
type curUserRootKey struct {
	privateKey []byte
	publicKey  []byte
}

// 获得当前用户的私钥和公钥
func getCurUserKey() (*curUserRootKey, error) {
	//读取本地 当前用户 根目录下的秘钥文件
	userInfo, _ := user.Current()
	userRootDir := userInfo.HomeDir
	//秘钥文件
	privateKeyFile := userRootDir + string(filepath.Separator) + ".ssh" + string(filepath.Separator) + "id_rsa"
	publicKeyFile := userRootDir + string(filepath.Separator) + ".ssh" + string(filepath.Separator) + "id_rsa.pub"

	_, err := os.Stat(privateKeyFile)
	if err != nil {
		return nil, errors.New(("获取当前用户下的私钥文件失败" + err.Error()))
	}

	_, err = os.Stat(publicKeyFile)
	if err != nil {
		return nil, errors.New(("获取当前用户下的公钥文件失败" + err.Error()))
	}

	priKey, _ := os.ReadFile(privateKeyFile)
	pubKey, _ := os.ReadFile(publicKeyFile)

	return &curUserRootKey{
		privateKey: priKey,
		publicKey:  pubKey,
	}, nil
}
