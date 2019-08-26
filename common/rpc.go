package common

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"qitmeer-miner/common/socks"
	"strings"
	"time"
)

type RpcClient struct {
	Cfg *GlobalConfig
}
// newHTTPClient returns a new HTTP client that is configured according to the
// proxy and TLS settings in the associated connection configuration.
func (rpc *RpcClient)newHTTPClient() (*http.Client, error) {
	// Configure proxy if needed.
	var dial func(network, addr string) (net.Conn, error)
	if rpc.Cfg.OptionConfig.Proxy != "" {
		proxy := &socks.Proxy{
			Addr:     rpc.Cfg.OptionConfig.Proxy,
			Username: rpc.Cfg.OptionConfig.ProxyUser,
			Password: rpc.Cfg.OptionConfig.ProxyPass,
		}
		dial = func(network, addr string) (net.Conn, error) {
			c, err := proxy.Dial(network, addr)
			if err != nil {
				return nil, err
			}
			return c, nil
		}
	}

	// Configure TLS if needed.
	var tlsConfig *tls.Config
	if !rpc.Cfg.SoloConfig.NoTLS && rpc.Cfg.SoloConfig.RPCCert != "" {
		pem, err := ioutil.ReadFile(rpc.Cfg.SoloConfig.RPCCert)
		if err != nil {
			return nil, err
		}

		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(pem)
		tlsConfig = &tls.Config{
			RootCAs:            pool,
			InsecureSkipVerify: rpc.Cfg.SoloConfig.NoTLS,
		}
	} else {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: rpc.Cfg.SoloConfig.NoTLS,
		}
	}

	// Create and return the new HTTP client potentially configured with a
	// proxy and TLS.
	client := http.Client{
		Transport: &http.Transport{
			Dial:            dial,
			TLSClientConfig: tlsConfig,
			DialContext: (&net.Dialer{
				Timeout:   time.Duration(rpc.Cfg.OptionConfig.Timeout) * time.Second,
				KeepAlive: time.Duration(rpc.Cfg.OptionConfig.Timeout) * time.Second,
				DualStack: true,
			}).DialContext,
		},
	}
	return &client, nil
}

func (rpc *RpcClient)RpcResult(method string,params []interface{}) []byte{
	protocol := "http"
	if !rpc.Cfg.SoloConfig.NoTLS {
		protocol = "https"
	}
	paramStr,err := json.Marshal(params)
	if err != nil {
		MinerLoger.Errorf("rpc params error:%v",err)
		return nil
	}
	url := rpc.Cfg.SoloConfig.RPCServer
	if !strings.Contains(rpc.Cfg.SoloConfig.RPCServer,"://"){
		url = protocol + "://" + url
	}
	jsonStr := []byte(`{"jsonrpc": "2.0", "method": "`+method+`", "params": `+string(paramStr)+`, "id": 1}`)
	bodyBuff := bytes.NewBuffer(jsonStr)
	httpRequest, err := http.NewRequest("POST", url, bodyBuff)
	if err != nil {
		MinerLoger.Errorf("rpc connect failed %v",err)
		return nil
	}
	httpRequest.Close = true
	httpRequest.Header.Set("Content-Type", "application/json")
	// Configure basic access authorization.
	httpRequest.SetBasicAuth(rpc.Cfg.SoloConfig.RPCUser, rpc.Cfg.SoloConfig.RPCPassword)

	// Create the new HTTP client that is configured according to the user-
	// specified options and submit the request.
	httpClient, err := rpc.newHTTPClient()
	if err != nil {
		MinerLoger.Errorf("rpc auth faild %v",err)
		return nil
	}
	httpClient.Timeout = time.Duration(rpc.Cfg.OptionConfig.Timeout) * time.Second
	httpResponse, err := httpClient.Do(httpRequest)

	if err != nil {
		MinerLoger.Errorf("rpc request faild %v",err)
		return nil
	}
	body, err := ioutil.ReadAll(httpResponse.Body)
	_ = httpResponse.Body.Close()
	if err != nil {
		MinerLoger.Errorf("error reading json reply:%v", err)
		return nil
	}

	if httpResponse.Status != "200 OK" {
		MinerLoger.Errorf("error http response :%v",  httpResponse.Status, string(body))
		return nil
	}
	return body
}
