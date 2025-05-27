package git

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
)

type GitConfig struct {
	RepoURL              string
	Branch               string
	FileName             string
	CommitAuthor         string
	CommitEmail          string
	AuthUsername         string
	AuthPassword         string
	AuthSSHKey           string
	AuthSSHKeyPassphrase string
}

type GitProvider struct {
	config GitConfig
	repo   *git.Repository
	auth   transport.AuthMethod
}

func NewGitProvider(config GitConfig) (*GitProvider, error) {
	slog.Debug("Creating Git provider", "repo_url", config.RepoURL, "branch", config.Branch)
	// 配置认证方式
	var auth transport.AuthMethod
	if config.AuthSSHKey != "" {
		// 使用 SSH 密钥认证
		sshKey, err := os.ReadFile(config.AuthSSHKey)
		if err != nil {
			return nil, fmt.Errorf("failed to read SSH key: %w", err)
		}
		auth, err = ssh.NewPublicKeys("git", sshKey, config.AuthSSHKeyPassphrase)
		if err != nil {
			return nil, fmt.Errorf("failed to create SSH auth: %w", err)
		}
	} else if config.AuthUsername != "" && config.AuthPassword != "" {
		// 使用用户名密码认证
		auth = &http.BasicAuth{
			Username: config.AuthUsername,
			Password: config.AuthPassword,
		}
	}

	// 克隆仓库
	repo, err := git.Clone(memory.NewStorage(), memfs.New(), &git.CloneOptions{
		URL:           config.RepoURL,
		Auth:          auth,
		SingleBranch:  true,
		ReferenceName: plumbing.NewBranchReferenceName(config.Branch),
		Depth:         1,
		NoCheckout:    false,
		Progress:      os.Stdout,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	// 检查分支是否存在
	_, err = repo.Reference(plumbing.NewBranchReferenceName(config.Branch), true)
	if err != nil {
		fmt.Println(1)
		// 如果分支不存在，创建新分支
		headRef, err := repo.Head()
		if err != nil {
			return nil, fmt.Errorf("failed to get HEAD reference: %w", err)
		}

		// 创建新分支
		branchRef := plumbing.NewHashReference(
			plumbing.NewBranchReferenceName(config.Branch),
			headRef.Hash(),
		)
		if err := repo.Storer.SetReference(branchRef); err != nil {
			return nil, fmt.Errorf("failed to create branch: %w", err)
		}

		// 切换到新分支
		worktree, err := repo.Worktree()
		if err != nil {
			return nil, fmt.Errorf("failed to get worktree: %w", err)
		}
		if err := worktree.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(config.Branch),
			Create: true,
		}); err != nil {
			return nil, fmt.Errorf("failed to checkout branch: %w", err)
		}
	}

	if config.CommitAuthor == "" {
		config.CommitAuthor = "ADIF2Cloud"
	}
	if config.CommitEmail == "" {
		config.CommitEmail = "adif2cloud@esd.cc"
	}

	return &GitProvider{
		config: config,
		repo:   repo,
		auth:   auth,
	}, nil
}

func (p *GitProvider) GetName() string {
	return fmt.Sprintf("Git->%s", p.config.RepoURL)
}

func (p *GitProvider) GetSize() (int64, error) {
	worktree, err := p.repo.Worktree()
	if err != nil {
		return 0, fmt.Errorf("failed to get worktree: %w", err)
	}

	fileInfo, err := worktree.Filesystem.Stat(p.config.FileName)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get file info: %w", err)
	}

	return fileInfo.Size(), nil
}

func (p *GitProvider) Download(w io.Writer) error {
	worktree, err := p.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	file, err := worktree.Filesystem.Open(p.config.FileName)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(w, file)
	return err
}

func (p *GitProvider) Upload(sourceFilePath string, _ string) error {
	worktree, err := p.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	err = worktree.Pull(&git.PullOptions{
		Auth: p.auth,
	})
	if err != nil {
		return fmt.Errorf("failed to pull: %w", err)
	}

	// 读取源文件
	sourceFile, err := os.Open(sourceFilePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	// 写到仓库
	repoFile, err := worktree.Filesystem.OpenFile(p.config.FileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer repoFile.Close()

	// 写文件
	io.Copy(repoFile, sourceFile)

	// 添加文件到暂存区
	if _, err := worktree.Add(p.config.FileName); err != nil {
		return fmt.Errorf("failed to add file: %w", err)
	}

	// 提交更改
	commit, err := worktree.Commit("Update ADIF file", &git.CommitOptions{
		Author: &object.Signature{
			Name:  p.config.CommitAuthor,
			Email: p.config.CommitEmail,
			When:  time.Now(),
		},
		SignKey: nil,
	})
	if err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	// 推送到远程仓库
	if err := p.repo.Push(&git.PushOptions{
		Auth: p.auth,
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("+%s:%s", plumbing.NewBranchReferenceName(p.config.Branch), plumbing.NewBranchReferenceName(p.config.Branch))),
		},
	}); err != nil {
		return fmt.Errorf("failed to push: %w", err)
	}

	slog.Info("Successfully pushed to git repository",
		"commit", commit.String(),
		"branch", p.config.Branch)
	return nil
}
