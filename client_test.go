package tapi

import (
	"crypto/hmac"
	"testing"
)

const (
	fakeID  = "b493d48364afe44d11c0165cf470a4164d1e2609911ef998be868d46ade3de4e"
	fakeKey = "1ebda7d457ece1330dff1c9e04cd62c4e02d1835968ff89d2fb2339f06f73028"
)

func TestClientHmac(t *testing.T) {
	expected := "7f59ea8749ba596d5c23fa242a531746b918e5e61c9f6c86" +
		"63a699736db503980f3a507ff7e2ef1336f7888d684a06c9a460d182" +
		"90e7b738a61d03e25ffdeb76"
	cli := NewClient(DefaultService, fakeID, fakeKey, nil)
	msg := "/tapi/v3/?tapi_method=list_orders&tapi_nonce=1"
	mac := cli.Hmac(msg)
	ok := hmac.Equal([]byte(mac), []byte(expected))
	if !ok {
		t.Errorf("wrong hmac")
	}
}
