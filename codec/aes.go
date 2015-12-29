package codec

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	//log "github.com/cihub/seelog"
)

const (
	//default_password = "medtree-im-passwmedtree-im-passw"
	default_password = "medtree-im-passw"
)

type AESCodec struct {
	Password []byte
	Iv       []byte
	block    cipher.Block
}

func NewAESCodec() (*AESCodec, error) {
	password := []byte(default_password)

	block, err := aes.NewCipher(password)
	if err != nil {
		return nil, err
	}

	codec := new(AESCodec)
	codec.Password = password
	codec.Iv = password[:16]
	codec.block = block

	return codec, nil
}

func (codec *AESCodec) ChangePassword(password string) error {
	pd := []byte(password)

	block, err := aes.NewCipher(pd)
	if err != nil {
		return err
	}

	codec.Password = pd
	codec.Iv = pd[:16]
	codec.block = block

	return nil
}

func (codec *AESCodec) Encrypt(src []byte) []byte {
	encrypter := cipher.NewCBCEncrypter(codec.block, codec.Iv)
	src = PKCS7Padding(src, codec.block.BlockSize())

	dst := make([]byte, len(src))
	encrypter.CryptBlocks(dst, src)

	return dst
}

func (codec *AESCodec) Decrypt(src []byte) []byte {
	decrypter := cipher.NewCBCDecrypter(codec.block, codec.Iv)

	dst := make([]byte, len(src))
	decrypter.CryptBlocks(dst, src)
	dst = PKCS7UnPadding(dst, codec.block.BlockSize())

	return dst
}

func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS7UnPadding(plantText []byte, blockSize int) []byte {
	length := len(plantText)
	unpadding := int(plantText[length-1])
	return plantText[:(length - unpadding)]
}
