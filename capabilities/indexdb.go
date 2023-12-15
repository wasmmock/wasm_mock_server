package capabilities

import (
	"github.com/wasmmock/wasm_mock_server/util"
)

func IndexDbStore(safeIndexDb *util.SafeIndexDb, binding string, req []byte) ([]byte, error) {
	safeIndexDb.Store(binding, req)
	return []byte{}, nil
}
func IndexDbGet(safeIndexDb *util.SafeIndexDb, binding string) ([]byte, error) {
	return safeIndexDb.Get(binding)
}
