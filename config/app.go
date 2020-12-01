package config

type App struct {
	Session struct {
		Name      string `yaml:"name"`
		SecretKey string `yaml:"secret_key"`
	} `yaml:"session"`
	Db struct {
		Default Db
	}
	Redis struct {
		Default Redis
	}
	OAuth2 struct {
		Client []Client `yaml:"client"`
	} `yaml:"oauth2"`
}

type Db struct {
	DriveName string `yaml:"drivename"`
	Host      string `yaml:"host"`
	Port      string    `yaml:"port"`
	User      string `yaml:"user"`
	Password  string `yaml:"password"`
	DbName    string `yaml:"dbname"`
}

type Redis struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	Db       int    `yaml:"db"`
}

type Client struct {
	ID     string  `yaml:"id"`
	Secret string  `yaml:"secret"`
	Name   string  `yaml:"name"`
	Domain string  `yaml:"domain"`
	Scope  []Scope `yaml:"scope"`
}

type Scope struct {
	ID    string `yaml:"id"`
	Title string `yaml:"title"`
}
