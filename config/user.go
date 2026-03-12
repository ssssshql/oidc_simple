package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

type User struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"` // 明文密码，通过文件权限保护
	Email    string `yaml:"email"`
	Name     string `yaml:"name"`
}

type Config struct {
	Users []User `yaml:"users"`
}

var (
	config     *Config
	configOnce sync.Once
	configErr  error
)

// LoadConfig 加载配置文件，如果不存在则自动创建默认配置
func LoadConfig(path string) (*Config, error) {
	// 如果已经成功加载过，直接返回
	if config != nil {
		return config, nil
	}

	var err error
	var cfg Config

	// 尝试读取文件
	data, err := os.ReadFile(path)
	if err != nil {
		// 文件不存在，创建默认配置
		if os.IsNotExist(err) {
			if err := createDefaultConfig(path); err != nil {
				return nil, fmt.Errorf("创建默认配置文件失败: %w", err)
			}
			// 重新读取刚创建的默认配置
			data, err = os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("读取默认配置文件失败: %w", err)
			}
		} else {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}
	}

	// 解析 YAML
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 保存到全局变量
	config = &cfg

	return config, nil
}

// createDefaultConfig 创建默认配置文件
func createDefaultConfig(path string) error {
	defaultConfig := `# OIDC 用户配置文件
# ⚠️  安全警告：这是自动生成的默认配置，请立即修改默认密码！
# ⚠️  请设置文件权限为 600: chmod 600 config.yml

users:
  - username: admin
    password: admin123
    email: admin@example.com
    name: 管理员
`

	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	// 写入配置文件（权限 600，仅所有者可读写）
	if err := os.WriteFile(path, []byte(defaultConfig), 0600); err != nil {
		return fmt.Errorf("写入默认配置文件失败: %w", err)
	}

	fmt.Printf("✓ 已自动创建默认配置文件: %s\n", path)
	fmt.Println("⚠️  请立即修改默认管理员密码！")

	return nil
}

// GetUser 根据用户名获取用户信息
func GetUser(username string) *User {
	if config == nil {
		fmt.Println("[DEBUG] config 是 nil！")
		return nil
	}

	fmt.Printf("[DEBUG] config.Users 长度: %d\n", len(config.Users))

	for i := range config.Users {
		fmt.Printf("[DEBUG] 检查用户: %s\n", config.Users[i].Username)
		if config.Users[i].Username == username {
			fmt.Printf("[DEBUG] 找到用户: %+v\n", config.Users[i])
			return &config.Users[i]
		}
	}

	fmt.Printf("[DEBUG] 未找到用户: %s\n", username)
	return nil
}

// ValidateUser 验证用户名和密码
func ValidateUser(username, password string) *User {
	fmt.Printf("[DEBUG] ValidateUser 被调用: %s / %s\n", username, password)

	user := GetUser(username)
	if user == nil {
		fmt.Println("[DEBUG] 用户不存在")
		return nil
	}

	// 明文密码验证
	if user.Password == password {
		fmt.Println("[DEBUG] 密码验证成功")
		return user
	}

	fmt.Printf("[DEBUG] 密码不匹配: 期望=%s, 实际=%s\n", user.Password, password)
	return nil
}
