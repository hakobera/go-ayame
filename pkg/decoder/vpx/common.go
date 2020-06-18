package vpx

// #cgo LDFLAGS: -lvpx
// #include <stdlib.h>
// #include "binding.h"
import "C"

import "fmt"

func handleError(ctx *C.vpx_codec_ctx_t, msg string) error {
	code := C.GoString(C.vpx_codec_error(ctx))
	detail := C.GoString(C.vpx_codec_error_detail(ctx))
	return fmt.Errorf("%s %s : %s", msg, code, detail)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
