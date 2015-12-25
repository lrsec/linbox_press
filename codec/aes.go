package codec

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
)

const (
	default_password = "medtree-im-passwmedtree-im-passw"
)

type AESCodec struct {
	Password  []byte
	block     cipher.Block
	encrypter cipher.BlockMode
	decrypter cipher.BlockMode
}

func NewAESCodec() (*AESCodec, error) {
	password := []byte(default_password)

	block, err := aes.NewCipher(password)
	if err != nil {
		return nil, err
	}

	iv := password[:16]

	codec := new(AESCodec)
	codec.Password = password
	codec.block = block
	codec.encrypter = cipher.NewCBCEncrypter(block, iv)
	codec.decrypter = cipher.NewCBCDecrypter(block, iv)

	return codec, nil
}

func (codec *AESCodec) ChangePassword(password string) error {
	pd := []byte(password)

	block, err := aes.NewCipher(pd)
	if err != nil {
		return err
	}

	length := len(password)

	iv := make([]byte, 8*length, 8*length)
	for i := 0; i < 8; i++ {
		copy(iv[i*length:(i+1)*length], password)
	}

	codec.Password = pd
	codec.block = block
	codec.encrypter = cipher.NewCBCEncrypter(block, iv)
	codec.decrypter = cipher.NewCBCDecrypter(block, iv)

	return nil
}

func (codec *AESCodec) Encrypt(src []byte) []byte {
	src = PKCS7Padding(src, codec.block.BlockSize())
	dst := make([]byte, len(src))
	codec.encrypter.CryptBlocks(dst, src)

	return dst
}

func (codec *AESCodec) Decrypt(src []byte) []byte {
	dst := make([]byte, len(src))
	codec.decrypter.CryptBlocks(dst, src)
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
