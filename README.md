# Wasm Mock Server
<img src="https://github.com/wasmmock/wasm_mock_rust/blob/main/hammock-min.png" width="100" height="100">

MITM server for software testing that supports both user defined wasm functions and WS connection.
https://hub.docker.com/repository/docker/rustropy/wasm_mock_server/

<img src="https://rustropy.netlify.app/images/wasmtesting.png" width="600" height="400">

![Alt Text](./desktop.gif)

## Getting Started
Head to https://github.com/wasmmock/wasm_mock_server/releases and download the relevant zip. Unzip the binary and run the binary.

* API default port: 20825
* API HTTP MITM default port: 20810 
* User defined port for TCP MITM: 3335 (for example)


## For starter
- Download v2.wasm from this repo.
The source code for this wasm can be found [here](https://github.com/wasmmock/wasm_mock_rust/blob/main/examples/http_fiddler/v2.rs)

```shell
curl -X POST http://localhost:20825/call/v2/unified \
    --header "Content-Type:application/octet-stream" \
	--data-binary "@v2.wasm"
```
- You should see "added_fn /t.json_http_modify_req,/t.json_http_modify_res"

## HTTPs Proxy
- SSL Cert can be found in GET /cert/pem
```shell
wget http://localhost:20825/cert/pem
```
- On Macos, click on the crt file to install the SSL certificate. 
### Firefox testing
- go to the firefox's certificate manager and import the CRT certificate and trust it.
- In search bar insdie firefox's setting, type "proxy"
- Check Manual proxy config box
    - Type your wifi ip address inside HTTP PROXY text box
    - Check the box "Also to use this proxy for HTTPS"
    - Type "20810" in Port
    - Click "Ok"
- In the firefox browser, go to www.yahoo.com/t.json. You should see
```json
{"data":"hi"}
```
### Mobile testing (Android)
Copy c8750f0d.0 from cert_folder from the zip files into /system/etc/security/cacerts folder in Android phone.

In order to do this, rooting the phone is required.
Please check this stackoverflow for detail: https://stackoverflow.com/questions/44942851/install-user-certificate-via-adb

## Video Resources
| Video Description  | Video Link |
| ------------- | ------------- |
| CNCF Sandbox Proposal Demo  | https://www.youtube.com/watch?v=Jte4n2pb5Y8&t=263s  |
| [Software testing] Wasm mock server websocket mitm (1)  | https://www.youtube.com/watch?v=xuspE_u71Og  |
| 【软件测试】wasm mock server Websocket 示范  | https://www.bilibili.com/video/BV1kg4y157hE/?spm_id_from=333.999.0.0&vd_source=8513215e56d2a613eb870e5ccc630e88  |
| 【rust conf china】应用WAPC做软件测试工具   | https://www.bilibili.com/video/BV1ws4y1k7pR/?spm_id_from=333.999.0.0&vd_source=8513215e56d2a613eb870e5ccc630e88  |

## Roadmap
| Task  | Completion | Details |
| ------------- | ------------- | ------------- |
| Consolidate all endpoints into one && remove target url parameter  | ✅ | https://github.com/wasmmock/wasm_mock_server/releases/tag/v0.1.2    |
| Wasm Playground  | ✅    | (2023-08-31) <ul><li>RLS code autocomplete</li><li> Tested on mobile browser (chrome) and desktop browser (chrome and firefox)</li></ul> |
| Web UI   | ✅  | (2023-08-08) Demo in CNCF sandbox proposal https://github.com/cncf/sandbox/issues/50 |
| Live deployment   | ✅   | (2023-08-31) https://pg.wasmmock.xyz |
| Update of Tech Blog   |    | @ https://rustropy.dev |
| Documentation Page   |    | In the future, there will be a standalone page for wasmmock |
| GUI support for tracing and breakpoint   |    | Currently charles proxy's features have low priority. In the future,  |
| Community Engagement   |    | Slack / Discord / X |
| Appium Intergration (POC)  |    | Include python SDK for mobile automation for android emulator |
| UIautomator Intergration Plus (POC) |    | Using UIautomator's xml and MITM data, predicts possible test cases for mobile application   |