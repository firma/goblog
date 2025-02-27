package library

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

var tenToAny = map[int]string{0: "0", 1: "1", 2: "2", 3: "3", 4: "4", 5: "5", 6: "6", 7: "7", 8: "8", 9: "9", 10: "a", 11: "b", 12: "c", 13: "d", 14: "e", 15: "f", 16: "g", 17: "h", 18: "i", 19: "j", 20: "k", 21: "l", 22: "m", 23: "n", 24: "o", 25: "p", 26: "q", 27: "r", 28: "s", 29: "t", 30: "u", 31: "v", 32: "w", 33: "x", 34: "y", 35: "z", 37: "A", 38: "B", 39: "C", 40: "D", 41: "E", 42: "F", 43: "G", 44: "H", 45: "I", 46: "J", 47: "K", 48: "L", 49: "M", 50: "N", 51: "O", 52: "P", 53: "Q", 54: "R", 55: "S", 56: "T", 57: "U", 58: "V", 59: "W", 60: "X", 61: "Y", 62: "Z", 63: ":", 64: ";", 65: "<", 66: "=", 67: ">", 68: "?", 69: "@", 70: "[", 71: "]", 72: "^", 73: "_", 74: "{", 75: "|", 76: "}"}

func DecimalToAny(num int64, n int) string {
	newNumStr := ""
	var remainder int
	var remainderString string
	for num != 0 {
		remainder = int(num % int64(n))
		if 76 > remainder && remainder > 9 {
			remainderString = tenToAny[remainder]
		} else {
			remainderString = strconv.Itoa(remainder)
		}
		newNumStr = remainderString + newNumStr
		num = num / int64(n)
	}
	return newNumStr
}

func GenerateRandNumber(length uint) uint {
	numberByteArray := [9]byte{1, 2, 3, 4, 5, 6, 7, 9}
	numberLength := len(numberByteArray)
	rand.Seed(time.Now().UnixNano())

	var stringBuilder strings.Builder
	for i := 0; uint(i) < length; i++ {
		fmt.Fprintf(&stringBuilder, "%d", numberByteArray[rand.Intn(numberLength)])
	}
	randomNumber, _ := strconv.ParseUint(stringBuilder.String(), 10, 0)
	return uint(randomNumber)
}

func Md5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func Md5Bytes(bytes []byte) string {
	h := md5.New()
	h.Write(bytes)
	return hex.EncodeToString(h.Sum(nil))
}

// VersionCompare 对比两个版本，a > b = 1; a < b = -1;
func VersionCompare(a string, b string) int {
	aSplit := strings.Split(a, ".")
	bSplit := strings.Split(b, ".")

	aLen := len(aSplit)
	bLen := len(bSplit)
	length := min(aLen, bLen)

	for i := 0; i < length; i++ {
		intA, _ := strconv.ParseInt(aSplit[i], 10, 32)
		intB, _ := strconv.ParseInt(bSplit[i], 10, 32)
		if intA > intB {
			return 1
		} else if intB > intA {
			return -1
		}
	}
	if aLen > bLen {
		return 1
	} else if bLen > aLen {
		return -1
	}
	return 0
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
