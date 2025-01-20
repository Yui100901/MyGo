package mqtt_utils

//
// @Author yfy2001
// @Date 2024/12/29 14 27
//

type MQTTConfiguration struct {
	ID       string `yaml:"id"`
	URL      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}
