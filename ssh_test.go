package yiigo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSSHKey(t *testing.T) {
	_, err := GenerateSSHKey()

	assert.Nil(t, err)

	// 结果已通过阿里云密钥对验证
	// --- id_rsa.pub ---
	// ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCePoas3sNOSeC8ODs+9hXSqug+Km9F4rezvDEUry2IqoYs1Jj46gUL+eLPP531rUIk+pK+ClVL/dd6jCwEILTLV3N1PGIW4vIE8Chw+6u1OmJOwZcj8t5InQds9nTGpgIpGOJWQSicrYmQS2b6OkCgWr9V95a2VamPJY/ZuRZ9ZdGyyNY7cizeQzOO/+5lN8kCbG3qGedh1SbPYU/RktE9fPh9xpWpQzFQiqvtiDEItj1SNhagaO8ua/tp3n47SeJo9kYFaJ5C/M06L9B+g/yzgESCAsSZaRvH0sc2E90RV+xL4qzMfuCOw9jDCNq7yO5ONuZSIlupxLnXY0BreY/X
	// --- key fingerprint ---
	// 75fa3a2754ac147637b79d4276885aa8
}

func TestRSAPemToSSH(t *testing.T) {
	rsaPubKey := `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEApAsNx7DxWcjyaRpNP2Vx
nFGkvtgpZwn77FBu7pbKU54YnsRXWor2+B2RBwscmlTjkCA46y1OTWAEBEAE1qP8
V2pZHC2Z7X07+PU4R4Rb28YzOXUdsdZMykGxr4YBbhQdrqgOzBJLeJY0f7siajjH
6eKn/raiLGWLAxzdZ3oMcdqShJsek1BTCFUT3IcwyCvGPVKTMTLMSzl6F8gxuVc4
YGCv8SHug7cjWo1H1kEx3kXpvNy3JcE2d/BEdQqPRzrfNueXHNpK/DNwNxcvMbZ9
u0X/QQGFgify3nPVfUw+eZhHPxmBGwWzMraa8qJ0s/mO1UXEfSjZIIvmpnSl2Xkv
CQIDAQAB
-----END PUBLIC KEY-----`

	sshRSA, fingerprint, err := RSAPemToSSH([]byte(rsaPubKey))

	assert.Nil(t, err)

	// 结果已通过阿里云密钥对验证
	assert.Equal(t, "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCkCw3HsPFZyPJpGk0/ZXGcUaS+2ClnCfvsUG7ulspTnhiexFdaivb4HZEHCxyaVOOQIDjrLU5NYAQEQATWo/xXalkcLZntfTv49ThHhFvbxjM5dR2x1kzKQbGvhgFuFB2uqA7MEkt4ljR/uyJqOMfp4qf+tqIsZYsDHN1negxx2pKEmx6TUFMIVRPchzDIK8Y9UpMxMsxLOXoXyDG5VzhgYK/xIe6DtyNajUfWQTHeRem83LclwTZ38ER1Co9HOt8255cc2kr8M3A3Fy8xtn27Rf9BAYWCJ/Lec9V9TD55mEc/GYEbBbMytpryonSz+Y7VRcR9KNkgi+amdKXZeS8J\n", string(sshRSA))
	assert.Equal(t, "2a1221695eecd9a88cf4b07f84ff40c0", fingerprint)
}
