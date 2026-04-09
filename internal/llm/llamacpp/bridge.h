#pragma once

#ifdef __cplusplus
extern "C" {
#endif

void * th_llama_model_load(const char * model_path, int n_ctx, int threads, int n_gpu_layers, char * err_buf, int err_buf_len);
int th_llama_model_predict(void * handle, const char * prompt, int max_tokens, char * out_buf, int out_buf_len, char * err_buf, int err_buf_len);
void th_llama_model_free(void * handle);

#ifdef __cplusplus
}
#endif