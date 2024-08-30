
////////////////////////////////
package misc

import (
    "fmt"
    "strconv"
    "encoding/hex"
    "golang.org/x/crypto/blake2b"
)

////////////////////////////////
// Convert the public key list to script hash of multisig.
func ConvKPubListToScriptHashMultisig(m int64, kPubList []string, n int64) (string) {
    lenKPubList := len(kPubList)
    if (lenKPubList < 1 || lenKPubList != int(n)) {
        return ""
    }
    ecdsa := false
    kPub := strconv.FormatInt(m+80, 16)
    for _, k := range kPubList {
        lenK := len(k)
        if lenK == 64 {
            kPub += "20"
        } else if lenK == 66 {
            kPub += "21"
            ecdsa = true
        } else {
            return ""
        }
        kPub += k
    }
    kPub += strconv.FormatInt(n+80, 16)
    if ecdsa {
        kPub += "a9"
    } else {
        kPub += "ae"
    }
    decoded, _ := hex.DecodeString(kPub)
    sum := blake2b.Sum256(decoded)
    return fmt.Sprintf("%064x", string(sum[:]))
}

////////////////////////////////
// Convert the script hash to address.
func ConvKPubToP2sh(kPub string, testnet bool) (string) {
    lenKey := len(kPub)
    if lenKey != 64 {
        return ""
    }
    kPub = "08" + kPub  // P2SH ver
    decoded, _ := hex.DecodeString(kPub)
    kPub = string(decoded[:])
    addr := EncodeBech32(kPub, testnet)
    if testnet {
        addr = "kaspatest:" + addr
    } else {
        addr = "kaspa:" + addr
    }
    return addr
}

////////////////////////////////
// Convert the public key to address.
func ConvKPubToAddr(kPub string, testnet bool) (string) {
    lenKey := len(kPub)
    if lenKey == 64 {  // Schnorr ver
        kPub = "00" + kPub
    } else if lenKey == 66 {  // Ecdsa ver
        kPub = "01" + kPub
    } else {
        return ""
    }
    decoded, _ := hex.DecodeString(kPub)
    kPub = string(decoded[:])
    addr := EncodeBech32(kPub, testnet)
    if testnet {
        addr = "kaspatest:" + addr
    } else {
        addr = "kaspa:" + addr
    }
    return addr
}

////////////////////////////////
// Convert the address to public key or script hash.
func ConvAddrToKPub(addr string, testnet bool) (string, string) {
    s := 6
    if testnet {
        s = 10
    }
    if (!testnet && (len(addr) < 67 || addr[0:s] != "kaspa:")) {
        return "", ""
    }
    if (testnet && (len(addr) < 71 || addr[0:s] != "kaspatest:")) {
        return "", ""
    }
    kPub := hex.EncodeToString([]byte(DecodeBech32(addr[s:], testnet)))
    if len(kPub) < 64 {
        return "", ""
    }
    return kPub[:2], kPub[2:]
}

////////////////////////////////
// Verify the address.
func VerifyAddr(addr string, testnet bool) (bool) {
    ver, kPub := ConvAddrToKPub(addr, testnet)
    if kPub == "" {
        return false
    }
    addr2 := ""
    if ver == "08" {
        addr2 = ConvKPubToP2sh(kPub, testnet)
    } else {
        addr2 = ConvKPubToAddr(kPub, testnet)
    }
    if addr2 != addr {
        return false
    }
    return true
}

////////////////////////////////
func EncodeBech32(data string, testnet bool) (string) {
    _pMod := func(list []byte) int {
        g := []int{0x98f2bc8e61, 0x79b76d99e2, 0xf33e5fb3c4, 0xae2eabe2a8, 0x1e4f43e470}
        cs := 1
        for _, v := range list {
            b := cs >> 35
            cs = ((cs & 0x07ffffffff) << 5) ^ int(v)
            for i := 0; i < len(g); i++ {
                if ((b >> uint(i)) & 1) == 1 {
                    cs ^= g[i]
                }
            }
        }
        return cs ^ 1
    }
    b8 := []byte(data)
    b8Len := len(b8)
    nLast := 0
    bLast := byte(0)
    var b5 []byte
    for i :=0; i < b8Len; i++ {
        rMove := 3 + nLast
        b := (b8[i] >> rMove) & 31
        if nLast > 0 {
            b |= bLast
        }
        b5 = append(b5, b)
        nLast = rMove
        if rMove >= 5 {
            b5 = append(b5, (b8[i] << (8-rMove) >> 3) & 31)
            nLast = rMove - 5
        }
        if nLast > 0 {
            bLast = (b8[i] << (8-nLast) >> 3) & 31
        }
    }
    if nLast > 0 {
        b5 = append(b5, bLast)
    }
    b5ex := []byte{11, 1, 19, 16, 1, 0}
    if testnet {
        b5ex = []byte{11, 1, 19, 16, 1, 20, 5, 19, 20, 0}
    }
    b5ex = append(b5ex, b5...)
    b5ex = append(b5ex, 0, 0, 0, 0, 0, 0, 0, 0)
    p := _pMod(b5ex)
    c := []string{"q", "p", "z", "r", "y", "9", "x", "8", "g", "f", "2", "t", "v", "d", "w", "0", "s", "3", "j", "n", "5", "4", "k", "h", "c", "e", "6", "m", "u", "a", "7", "l"}
    for i := 0; i < 8; i++ {
        b5 = append(b5, byte((p >> (5*(7-i))) & 31))
    }
    result := ""
    for i := 0; i < len(b5); i++ {
        result += c[int(b5[i])]
    }
    return result
}

////////////////////////////////
func DecodeBech32(data string, testnet bool) (string) {
    _pMod := func(list []byte) int {
        g := []int{0x98f2bc8e61, 0x79b76d99e2, 0xf33e5fb3c4, 0xae2eabe2a8, 0x1e4f43e470}
        cs := 1
        for _, v := range list {
            b := cs >> 35
            cs = ((cs & 0x07ffffffff) << 5) ^ int(v)
            for i := 0; i < len(g); i++ {
                if ((b >> uint(i)) & 1) == 1 {
                    cs ^= g[i]
                }
            }
        }
        return cs ^ 1
    }
    n := map[string]byte{"q":0, "p":1, "z":2, "r":3, "y":4, "9":5, "x":6, "8":7, "g":8, "f":9, "2":10, "t":11, "v":12, "d":13, "w":14, "0":15, "s":16, "3":17, "j":18, "n":19, "5":20, "4":21, "k":22, "h":23, "c":24, "e":25, "6":26, "m":27, "u":28, "a":29, "7":30, "l":31}
    dataLen := len(data)
    var b5 []byte
    for i :=0; i < dataLen; i++ {
        _, existed := n[data[i:i+1]]
        if !existed {
            return ""
        }
        b5 = append(b5, n[data[i:i+1]])
    }
    b5Len := len(b5)
    cs := b5[b5Len-8:]
    b5 = b5[:b5Len-8]
    b5Len -= 8
    b5ex := []byte{11, 1, 19, 16, 1, 0}
    if testnet {
        b5ex = []byte{11, 1, 19, 16, 1, 20, 5, 19, 20, 0}
    }
    b5ex = append(b5ex, b5...)
    b5ex = append(b5ex, 0, 0, 0, 0, 0, 0, 0, 0)
    p := _pMod(b5ex)
    for i := 0; i < 8; i++ {
        if cs[i] != byte((p >> (5*(7-i))) & 31) {
            return ""
        }
    }
    var b8 []byte
    nLast := 0
    bLast := byte(0)
    for i :=0; i < b5Len; i++ {
        offset := 3 - nLast
        if offset == 0 {
            b8 = append(b8, b5[i] | bLast)
            nLast = 0
            bLast = 0
        } else if offset < 0 {
            b8 = append(b8, (b5[i] >> (-offset)) | bLast)
            nLast = -offset
            bLast = b5[i] << (8-nLast)
        } else {
            bLast |= b5[i] << offset
            nLast += 5
        }
    }
    return string(b8)
}
