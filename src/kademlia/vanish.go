package kademlia

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	mathrand "math/rand"
	"sss"
	"time"
)

type VanashingDataObject struct {
	AccessKey  int64
	Ciphertext []byte
	NumberKeys byte
	Threshold  byte
}

func GenerateRandomCryptoKey() (ret []byte) {
	for i := 0; i < 32; i++ {
		ret = append(ret, uint8(mathrand.Intn(256)))
	}
	return
}

func GenerateRandomAccessKey() (accessKey int64) {
	r := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
	accessKey = r.Int63()
	return
}

func CalculateSharedKeyLocations(accessKey int64, count int64) (ids []ID) {
	r := mathrand.New(mathrand.NewSource(accessKey))
	ids = make([]ID, count)
	for i := int64(0); i < count; i++ {
		for j := 0; j < IDBytes; j++ {
			ids[i][j] = uint8(r.Intn(256))
		}
	}
	return
}

func encrypt(key []byte, text []byte) (ciphertext []byte) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	ciphertext = make([]byte, aes.BlockSize+len(text))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], text)
	return
}

func decrypt(key []byte, ciphertext []byte) (text []byte) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	if len(ciphertext) < aes.BlockSize {
		panic("ciphertext is not long enough")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)
	return ciphertext
}

func VanishData(kadem Kademlia, data []byte, numberKeys byte,
	threshold byte) (vdo VanashingDataObject) {
	key := GenerateRandomCryptoKey()
	ciphertext := encrypt(key, data)
	shares, err := sss.Split(numberKeys, threshold, key)
	if err != nil {
		panic(err)
	}
	fullShares := sharesMap2Array(shares)
	accessKey := GenerateRandomAccessKey()
	locations := CalculateSharedKeyLocations(accessKey, int64(numberKeys))
	for i := byte(0); i < numberKeys; i++ {
		kadem.DoIterativeStore(locations[i], fullShares[i])
	}
	vdo.AccessKey = accessKey
	vdo.Ciphertext = ciphertext
	vdo.NumberKeys = numberKeys
	vdo.Threshold = threshold
	return
}

func UnvanishData(kadem Kademlia, vdo VanashingDataObject) (data []byte) {
	locations := CalculateSharedKeyLocations(vdo.AccessKey, int64(vdo.NumberKeys))
	threshold := int(vdo.Threshold)
	fullShares := make([][]byte, 0)
	for _, each := range locations {
		_, value := kadem.iterativeFindValue(each)
		if value != "" {
			fullShares = append(fullShares, []byte(value))
		}
		if len(fullShares) >= threshold {
			break
		}
	}
	if len(fullShares) < threshold {
		//panic("# of found shares less then threshold")
		data = nil
		return
	}
	shares := sharesArray2Map(fullShares)
	key := sss.Combine(shares)
	data = decrypt(key, vdo.Ciphertext)
	return
}

func sharesMap2Array(shares map[byte][]byte) (result [][]byte) {
	result = make([][]byte, 0)
	for k, v := range shares {
		all := append([]byte{k}, v...)
		result = append(result, all)
	}
	return
}

func sharesArray2Map(shares [][]byte) (result map[byte][]byte) {
	result = make(map[byte][]byte, 0)
	for _, all := range shares {
		k := all[0]
		v := all[1:]
		result[k] = v
	}
	return
}
