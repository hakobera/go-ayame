#ifndef __VPX_BINDING_H__
#define __VPX_BINDING_H__

#include <stdint.h>
#include <vpx/vpx_decoder.h>
#include <vpx/vp8dx.h>

extern vpx_codec_iface_t *ifaceVP8();
extern vpx_codec_iface_t *ifaceVP9();
extern vpx_codec_ctx_t *newVpxCtx();
extern vpx_codec_dec_cfg_t* newVpxDecCfg();
extern vp8_postproc_cfg_t* newVP8PostProcCfg();
extern vpx_codec_err_t vpxCodecControl(vpx_codec_ctx_t* ctx, int ctrl_id, void* data);

extern void yuv420_to_rgb(uint32_t width, uint32_t height,
                          const uint8_t *y, const uint8_t *u, const uint8_t *v,
                          int ystride, int ustride, int vstride,
                          uint8_t *out, int fmt);

#endif // __VPX_BINDING_H__