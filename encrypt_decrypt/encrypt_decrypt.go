package encrypt_decrypt

// XorEncryptDecrypt 异或加解密
func XorEncryptDecrypt(input []byte, key []byte) {
	if nil == input || 0 >= len(input) || nil == key || 0 >= len(key) {
		return
	}
	keyIndex := 0
	for index := 0; index < len(input); index++ {
		input[index] ^= key[keyIndex]
		keyIndex = (keyIndex + 1) % len(key)
	}
}
