package config

type TomlConfig struct {
	Log             Log            `toml:"log"`
	Token           Token          `toml:"token"`
	Http            HttpConfig     `toml:"http"`
	Https           HttpsConfig    `toml:"https"`
	Auto_cert       AutoCert       `toml:"auto_cert"`
	Redis           Redis          `toml:"redis"`
	Spr             Spr            `tome:"spr"`
	General_counter GeneralCounter `tome:"general_counter"`
	Db              DB             `toml:"db"`
	Elastic_search  ElasticSearch  `toml:"elastic_search"`
	Geo_ip          GeoIp          `toml:"geo_ip"`
	Leveldb         LevelDB        `toml:"leveldb"`
	Smtp            SMTP           `toml:"smtp"`
	Sqlite          Sqlite         `toml:"sqlite"`
}

type Log struct {
	Level string `toml:"level"`
}

type Token struct {
	Salt string `toml:"salt"`
}

type HttpConfig struct {
	Enable bool `toml:"enable"`
	Port   int  `toml:"port"`
}

type HttpsConfig struct {
	Enable   bool   `toml:"enable"`
	Port     int    `toml:"port"`
	Crt_path string `toml:"crt_path" `
	Key_path string `toml:"key_path"`
	Html_dir string `toml:"html_dir"`
}

type AutoCert struct {
	Enable         bool   `toml:"enable"`
	Check_interval int    `toml:"check_interval"`
	Crt_path       string `toml:"crt_path"`
	Init_download  bool   `toml:"init_download"`
	Key_path       string `toml:"key_path"`
	Url            string `toml:"url"`
}

type Spr struct {
	Enable bool `toml:"enable"`
}

type GeneralCounter struct {
	Enable       bool   `toml:"enable"`
	Project_name string `json:"project_name"`
}

type Redis struct {
	Enable   bool   `toml:"enable"`
	Use_tls  bool   `toml:"use_tls"`
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	Username string `toml:"username"`
	Password string `toml:"password"`
	Prefix   string `toml:"prefix"`
}

type DB struct {
	Enable   bool   `toml:"enable"`
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	Name     string `toml:"name"`
	Username string `toml:"username"`
	Password string `toml:"password"`
}

type ElasticSearch struct {
	Enable   bool   `toml:"enable"`
	Host     string `toml:"host"`
	Username string `toml:"username"`
	Password string `toml:"password"`
}

type GeoIp struct {
	Enable          bool   `toml:"enable"`
	Dataset_folder  string `toml:"dataset_folder"`
	Dataset_version string `toml:"dataset_version"`
}

type LevelDB struct {
	Enable bool   `toml:"enable"`
	Path   string `toml:"path"`
}

type SMTP struct {
	Enable     bool   `toml:"enable"`
	From_email string `toml:"from_email"`
	Host       string `toml:"host"`
	Port       int    `toml:"port"`
	Password   string `toml:"password"`
	Username   string `toml:"username"`
}

type Sqlite struct {
	Enable bool   `toml:"enable"`
	Path   string `toml:"path"`
}
