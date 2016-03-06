package mysql

import (
	"fmt"
	"database/sql"
	"github.com/kelseyhightower/confd/log"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

import _ "github.com/go-sql-driver/mysql"

// Client is a wrapper around the redis client
type Client struct {
	client   sql.DB
	machines []string
	password string
}

// Iterate through `machines`, trying to connect to each in turn.
// Returns the first successful connection or the last error encountered.
// Assumes that `machines` is non-empty.
func tryConnect(machines []string, password string) (sql.DB, error) {
	var err error
	log.Debug(fmt.Sprintf("Trying to connect to mysql node"))
		
	conn, err := sql.Open("mysql", "tmpl:iopiop@tcp(172.16.12.2:3306)/tmpl")
	if err != nil {
		panic(err.Error())
	}
	return *conn, nil
}

// Retrieves a connected redis client from the client wrapper.
// Existing connections will be tested with a PING command before being returned. Tries to reconnect once if necessary.
// Returns the established redis connection or the error encountered.
func (c *Client) connectedClient() (sql.DB, error) {
	return c.client, nil
}

// NewMySQLClient returns an *mysql.Client with a connection to named machines.
// It returns an error if a connection to the cluster cannot be made.
func NewMySQLClient(machines []string, password string) (*Client, error) {
	log.Info(fmt.Sprintf("NewMySQLClient"))
	var err error
	clientWrapper := &Client{ machines : machines, password: password }
	clientWrapper.client, err = tryConnect(machines, password)
	return clientWrapper, err
}

func Unpad(src []byte) []byte {
	length := len(src)
	unpadding := int(src[length-1])
	return src[:(length - unpadding)]
}

func decryptCBC(key []byte, cryptoText string) string {
	var block cipher.Block
	var err error

	//log.Info(fmt.Sprintf("decrypt [%s]", cryptoText))
	ciphertext, _ := base64.StdEncoding.DecodeString(cryptoText)

	padded_key := make([]byte, 32)
	copy(padded_key, []byte(key))

	if block, err = aes.NewCipher(padded_key); err != nil {
		panic("cannot create cipher")
	}

	if len(ciphertext) < aes.BlockSize {
		panic("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	cbc := cipher.NewCBCDecrypter(block, iv)
	cbc.CryptBlocks(ciphertext, ciphertext)

	plaintext := fmt.Sprintf("%s", Unpad(ciphertext))

	//log.Info(fmt.Sprintf("decrypted:[%s]", plaintext))
	return plaintext
}

// GetValues queries redis for keys prefixed by prefix.
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	vars := make(map[string]string)
	aes_key := []byte("HELP ME !!")

	for _, key := range keys {
		log.Info(fmt.Sprintf("GetValues [%s]", key))
		rows, err := c.client.Query("SELECT * FROM kv WHERE k LIKE '"+key+"%'")
		if err != nil {
			panic(err.Error())
		}
		for rows.Next() {
			var env int
			var version int
			var k string
			var v string
			rows.Scan(&env, &version, &k, &v)
			d := decryptCBC(aes_key, v)
			fmt.Printf("%s = %s\n", k, d)
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
