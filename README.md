# Wasm Mock Server
<img src="https://github.com/wasmmock/wasm_mock_rust/blob/main/hammock-min.png" width="100" height="100">

MITM server for software testing that supports both user defined wasm functions and WS connection.
https://hub.docker.com/repository/docker/rustropy/wasm_mock_server/

<img src="https://rustropy.netlify.app/images/wasmtesting.png" width="600" height="400">

## Getting Started
docker pull rustropy/wasm_mock_server:0.1.0

docker run -p 20825:20825 -p 20810:20810 -p 3335:3335 wasm_mock_server

* API default port: 20825
* API HTTP MITM default port: 20810 
* User defined port for TCP MITM: 3335 (for example)

## HTTPs Proxy
SSL Cert can be found in GET /cert/pem

## Video Resources
| Video Description  | Video Link |
| ------------- | ------------- |
| [Software testing] Wasm mock server websocket mitm (1)  | https://www.youtube.com/watch?v=xuspE_u71Og  |
| 【软件测试】wasm mock server Websocket 示范  | https://www.bilibili.com/video/BV1kg4y157hE/?spm_id_from=333.999.0.0&vd_source=8513215e56d2a613eb870e5ccc630e88  |
| 【rust conf china】应用WAPC做软件测试工具   | https://www.bilibili.com/video/BV1ws4y1k7pR/?spm_id_from=333.999.0.0&vd_source=8513215e56d2a613eb870e5ccc630e88  |

## Roadmap
| Task  | Completion | Detail |
| ------------- | ------------- | ------------- |
| Consolidate all endpoints into one && remove target url parameter  |  |
| Wasm Playground  | :heavy_check_mark:    |  |
| Web UI   |   |  |
| Live deployment   |   |  |