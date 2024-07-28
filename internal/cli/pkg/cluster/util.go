package cluster

import (
	"crypto/rand"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
	apitypes "github.com/oceanbase/ob-operator/api/types"

	"github.com/oceanbase/ob-operator/internal/dashboard/model/common"

	param "github.com/oceanbase/ob-operator/internal/dashboard/model/param"
)

func generateUserSecrets(clusterName string, clusterId int64) *apitypes.OBUserSecrets {
	return &apitypes.OBUserSecrets{
		Root:     fmt.Sprintf("%s-%d-root-%s", clusterName, clusterId, generateUUID()),
		ProxyRO:  fmt.Sprintf("%s-%d-proxyro-%s", clusterName, clusterId, generateUUID()),
		Monitor:  fmt.Sprintf("%s-%d-monitor-%s", clusterName, clusterId, generateUUID()),
		Operator: fmt.Sprintf("%s-%d-operator-%s", clusterName, clusterId, generateUUID()),
	}
}

func CheckResourceName(name string) bool {
	// 定义正则表达式
	regex := `[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*`

	// 编译正则表达式
	re, err := regexp.Compile(regex)
	if err != nil {
		fmt.Println("Error compiling regex:", err)
		return false
	}

	// 检查整个字符串是否符合正则表达式的模式
	return re.MatchString(name)
}
func CheckPassword(password string) bool {
	var (
		countUppercase   int
		countLowercase   int
		countNumber      int
		countSpecialChar int
	)

	// 遍历密码中的每个字符
	for _, char := range password {
		switch {
		case regexp.MustCompile(`[A-Z]`).MatchString(string(char)):
			countUppercase++
		case regexp.MustCompile(`[a-z]`).MatchString(string(char)):
			countLowercase++
		case regexp.MustCompile(`[0-9]`).MatchString(string(char)):
			countNumber++
		default:
			countSpecialChar++
		}
		// 提前返回
		if countUppercase >= 2 && countLowercase >= 2 && countNumber >= 2 && countSpecialChar >= 2 {
			return true
		}
	}
	return countUppercase >= 2 && countLowercase >= 2 && countNumber >= 2 && countSpecialChar >= 2
}
func mapZonesToTopology(zones map[string]string) ([]param.ZoneTopology, error) {
	if zones == nil {
		return nil, fmt.Errorf("Zone value is required") // 无效的zone信息
	}
	topology := make([]param.ZoneTopology, 0)
	for zoneName, replicaStr := range zones {
		replica, err := strconv.Atoi(replicaStr)
		if err != nil {
			return nil, fmt.Errorf("invalid value for zone %s: %s", zoneName, replicaStr)
		}
		topology = append(topology, param.ZoneTopology{
			Zone:         zoneName,
			Replicas:     replica,
			NodeSelector: make([]common.KVPair, 0),
			Tolerations:  make([]common.KVPair, 0),
			Affinities:   make([]common.AffinitySpec, 0),
		})
	}
	return topology, nil
}

func generateRandomPassword() string {
	const (
		maxLength      = 32
		minUppercase   = 2
		minLowercase   = 2
		minNumber      = 2
		minSpecialChar = 2
	)

	characters := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789~!@#%^&*_-+=|(){}[]:;,.?/`$\"<>"
	var (
		countUppercase   int
		countLowercase   int
		countNumber      int
		countSpecialChar int
	)

	var sb strings.Builder
	for countUppercase < minUppercase || countLowercase < minLowercase || countNumber < minNumber || countSpecialChar < minSpecialChar {
		b := make([]byte, 1)
		_, err := rand.Read(b) // 随机读取一个字节
		if err != nil {
			panic(err) // 处理错误
		}

		randomIndex := int(b[0]) % len(characters)
		randomChar := characters[randomIndex]

		sb.WriteByte(randomChar) // 追加字符到密码

		switch {
		case regexp.MustCompile(`[A-Z]`).MatchString(string(randomChar)):
			countUppercase++
		case regexp.MustCompile(`[a-z]`).MatchString(string(randomChar)):
			countLowercase++
		case regexp.MustCompile(`[0-9]`).MatchString(string(randomChar)):
			countNumber++
		default:
			countSpecialChar++
		}
		if len(sb.String()) >= maxLength {
			return generateRandomPassword()
		}
	}

	return sb.String()
}
func generateUUID() string {
	parts := strings.Split(uuid.New().String(), "-")
	return parts[len(parts)-1]
}
