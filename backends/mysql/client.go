package mysql

import (
	"fmt"
	"database/sql"
	"github.com/kelseyhightower/confd/log"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"os"
)

import _ "github.com/go-sql-driver/mysql"

// Client is a wrapper around the redis client
type Client struct {
	client              sql.DB
	backend_config_file string
	version             string
	key                 string
}

type MySQLConfig struct {
    Host string
    Name string
    User string
    Pass string
    Key string
}


// NewMySQLClient returns an *mysql.Client with a connection to named machines.
// It returns an error if a connection to the cluster cannot be made.
func NewMySQLClient(backend_config_file string, version string) (*Client, error) {
	log.Info(fmt.Sprintf("NewMySQLClient"))
	var err error
	clientWrapper := &Client{ backend_config_file : backend_config_file, version: version }

	log.Debug(fmt.Sprintf("backend_config_file: [%s]", backend_config_file))
	file, _ := os.Open(backend_config_file)
	decoder := json.NewDecoder(file)
	conf    := MySQLConfig{}
	err     = decoder.Decode(&conf)
	if err != nil {
		fmt.Println("error:", err)
	}

	dsn := conf.User+":"+conf.Pass+"@tcp("+conf.Host+":3306)/"+conf.Name
	log.Debug(fmt.Sprintf("Trying to connect to mysql: [%s]", dsn))
	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err.Error())
	}

	clientWrapper.key = conf.Key

	clientWrapper.client = *conn

	return clientWrapper, err
}

func Unpad(src []byte) []byte {
	length := len(src)
	unpadding := int(src[length-1])
	return src[:(length - unpadding)]
}

func decryptCBC(key []byte, cryptoText string, ivb64 string) string {
	var block cipher.Block
	var err error

	// log.Info(fmt.Sprintf("decrypt:[%s] iv:[%s]", cryptoText, ivb64))
	iv, _ := base64.StdEncoding.DecodeString(ivb64)
	ciphertext, _ := base64.StdEncoding.DecodeString(cryptoText)

	padded_key := make([]byte, 32)
	copy(padded_key, []byte(key))

	if block, err = aes.NewCipher(padded_key); err != nil {
		panic("cannot create cipher")
	}

	if len(ciphertext) < aes.BlockSize {
		panic("ciphertext too short")
	}

	cbc := cipher.NewCBCDecrypter(block, iv)
	cbc.CryptBlocks(ciphertext, ciphertext)

	plaintext := fmt.Sprintf("%s", Unpad(ciphertext))

	return plaintext
}

// GetValues queries redis for keys prefixed by prefix.
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	vars := make(map[string]string)

	aes_key := make([]byte, 32)
	copy(aes_key, c.key)

	for _, key := range keys {
		log.Info(fmt.Sprintf("GetValues [%s]", key))
		q := "SELECT * FROM config WHERE k LIKE '"+key+"%' AND version = "+c.version
		fmt.Printf("query:[%s]\n", q)
		rows, err := c.client.Query(q)
		if err != nil {
			panic(err.Error())
		}
		for rows.Next() {
			var environment_name string
			var id int
			var version int
			var k string
			var v string
			var iv string
			var d string
			var created string
			rows.Scan(&id, &environment_name, &version, &k, &v, &iv, &created)
			if v != "" && iv != "" {
				d = decryptCBC(aes_key, v, iv)
			} else {
				d = v
			}
			//fmt.Printf("%s = %s / %s\n", k, v, d)
			vars[k] = d
		}
	}
	return vars, nil
}

// WatchPrefix is not yet implemented.
func (c *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	<-stopChan
	return 0, nil
}
