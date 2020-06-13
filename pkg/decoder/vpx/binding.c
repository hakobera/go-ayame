#include <stdlib.h>
#include <vpx/vpx_image.h>

#include "binding.h"

vpx_codec_iface_t *ifaceVP8()
{
    return vpx_codec_vp8_dx();
}

vpx_codec_iface_t *ifaceVP9()
{
    return vpx_codec_vp9_dx();
}

vpx_codec_ctx_t *newVpxCtx()
{
    return malloc(sizeof(vpx_codec_ctx_t));
}

vpx_codec_dec_cfg_t *newVpxDecCfg()
{
    return malloc(sizeof(vpx_codec_dec_cfg_t));
}

vp8_postproc_cfg_t *newVP8PostProcCfg()
{
    return malloc(sizeof(vp8_postproc_cfg_t));
}

vpx_codec_err_t vpxCodecControl(vpx_codec_ctx_t *ctx, int ctrl_id, void *data)
{
    return vpx_codec_control_(ctx, ctrl_id, data);
}

void yuv420_to_rgb(uint32_t width, uint32_t height,
                   const uint8_t *y, const uint8_t *u, const uint8_t *v,
                   int ystride, int ustride, int vstride,
                   uint8_t *out, int fmt)
{
    unsigned long int i, j;
    for (i = 0; i < height; ++i)
    {
        for (j = 0; j < width; ++j)
        {
            uint8_t *point = out + 4 * ((i * width) + j);
            int t_y = y[((i * ystride) + j)];
            int t_u = u[(((i / 2) * ustride) + (j / 2))];
            int t_v = v[(((i / 2) * vstride) + (j / 2))];
            t_y = t_y < 16 ? 16 : t_y;
            int r = (298 * (t_y - 16) + 409 * (t_v - 128) + 128) >> 8;
            int g = (298 * (t_y - 16) - 100 * (t_u - 128) - 208 * (t_v - 128) + 128) >> 8;
            int b = (298 * (t_y - 16) + 516 * (t_u - 128) + 128) >> 8;
            if (fmt == 0 /* RGBA */)
            {
                point[0] = r > 255 ? 255 : r < 0 ? 0 : r;
                point[1] = g > 255 ? 255 : g < 0 ? 0 : g;
                point[2] = b > 255 ? 255 : b < 0 ? 0 : b;
                point[3] = ~0;
            }
            else
            { /* BGRA */
                point[0] = b > 255 ? 255 : b < 0 ? 0 : b;
                point[1] = g > 255 ? 255 : g < 0 ? 0 : g;
                point[2] = r > 255 ? 255 : r < 0 ? 0 : r;
                point[3] = ~0;
            }
        }
    }
}

uint8_t* newFrameBuffer(size_t n)
{
	return malloc(sizeof(uint8_t) * n);
}

static int32_t vpxGetFrameBuffer(void* user_priv, size_t min_size, vpx_codec_frame_buffer_t* fb)
{
	return goVpxGetFrameBuffer(user_priv, min_size, fb);
}

static int32_t vpxReleaseFrameBuffer(void* user_priv, vpx_codec_frame_buffer_t* fb)
{
	return goVpxReleaseFrameBuffer(user_priv, fb);
}

vpx_codec_err_t vpxCodecSetFrameBufferFunction(vpx_codec_ctx_t* ctx, void* user_priv)
{
    return vpx_codec_set_frame_buffer_functions(
        ctx,
        &vpxGetFrameBuffer,
        &vpxReleaseFrameBuffer,
        user_priv
    );
}
