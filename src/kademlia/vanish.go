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
	Timeout    byte
}

const EPOCH_RANGE = int64(3600 * 8)

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

func CalculateSharedKeyLocations(accessKey, count, epoch int64) (ids []ID) {
	r := mathrand.New(mathrand.NewSource(accessKey + epoch))
	ids = make([]ID, count)
	for i := int64(0); i < count; i++ {
		for j := 0; j < IDBytes; j++ {
			ids[i][j] = uint8(r.Intn(256))
		}
	}
	return
}

func currentEpoch() int64 {
	return time.Now().Unix() / EPOCH_RANGE
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

func VanishData(kadem Kademlia, data []byte, numberKeys, threshold, timeout byte) (vdo VanashingDataObject) {
	key := GenerateRandomCryptoKey()
	accessKey := GenerateRandomAccessKey()
	ciphertext := encrypt(key, data)
	distributeShares(kadem, numberKeys, threshold, key, accessKey)
	vdo.AccessKey = accessKey
	vdo.Ciphertext = ciphertext
	vdo.NumberKeys = numberKeys
	vdo.Threshold = threshold
	vdo.Timeout = timeout
	return
}

func distributeShares(kadem Kademlia, numberKeys, threshold byte, key []byte, accessKey int64) {
	shares, err := sss.Split(numberKeys, threshold, key)
	if err != nil {
		panic(err)
	}
	fullShares := sharesMap2Array(shares)
	locations := CalculateSharedKeyLocations(
		accessKey,
		int64(numberKeys),
		currentEpoch(),
	)
	for i := byte(0); i < numberKeys; i++ {
		kadem.DoIterativeStore(locations[i], fullShares[i])
	}
}

func Refresh(kadem Kademlia, vdo VanashingDataObject) {
	loops := int(vdo.Timeout) / 8
	for loops > 0 {
		select {
		case <-time.After(time.Hour * 8):
			key := retrieveKey(kadem, vdo)
			if key != nil {
				distributeShares(kadem, vdo.NumberKeys, vdo.Threshold, key, vdo.AccessKey)
			}
		}
	}
}

func UnvanishData(kadem Kademlia, vdo VanashingDataObject) (data []byte) {
	key := retrieveKey(kadem, vdo)
	if key == nil {
		data = nil
	} else {
		data = decrypt(key, vdo.Ciphertext)
	}
	return
}

func retrieveKey(kadem Kademlia, vdo VanashingDataObject) (key []byte) {
	threshold := int(vdo.Threshold)
	var fullShares [][]byte
	for i := range []int{0, -1, 1} {
		locations := CalculateSharedKeyLocations(
			vdo.AccessKey,
			int64(vdo.NumberKeys),
			currentEpoch()+int64(i),
		)
		fullShares = make([][]byte, 0)
		for _, each := range locations {
			_, value := kadem.iterativeFindValue(each)
			if value != "" {
				fullShares = append(fullShares, []byte(value))
			}
			if len(fullShares) >= threshold {
				break
			}
		}
		if len(fullShares) >= threshold {
			break
		}
	}
	if len(fullShares) < threshold {
		key = nil
		return
	}
	shares := sharesArray2Map(fullShares)
	key = sss.Combine(shares)
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
