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
extern vpx_codec_err_t vpxCodecSetFrameBufferFunction(vpx_codec_ctx_t* ctx, void* user_priv);

extern void yuv420_to_rgb(uint32_t width, uint32_t height,
                          const uint8_t *y, const uint8_t *u, const uint8_t *v,
                          int ystride, int ustride, int vstride,
                          uint8_t *out, int fmt);

extern uint8_t* newFrameBuffer(size_t n);

extern int32_t goVpxGetFrameBuffer(void*, size_t, vpx_codec_frame_buffer_t*);
extern int32_t goVpxReleaseFrameBuffer(void*, vpx_codec_frame_buffer_t*);

#endif // __VPX_BINDING_H__