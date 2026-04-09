#include "bridge.h"

#include "llama.h"

#include <algorithm>
#include <cstring>
#include <mutex>
#include <string>
#include <vector>

struct th_llama_model {
    llama_model * model;
    int n_ctx;
    int threads;
};

static void th_copy_error(char * err_buf, int err_buf_len, const std::string & message) {
    if (err_buf == nullptr || err_buf_len <= 0) {
        return;
    }

    const int copy_len = std::min<int>(static_cast<int>(message.size()), err_buf_len - 1);
    std::memcpy(err_buf, message.data(), static_cast<size_t>(copy_len));
    err_buf[copy_len] = '\0';
}

static void th_init_backend() {
    static std::once_flag once;
    std::call_once(once, []() {
        llama_log_set([](enum ggml_log_level /* level */, const char * /* text */, void * /* user_data */) {}, nullptr);

        llama_backend_init();
        ggml_backend_load_all();
    });
}

void * th_llama_model_load(const char * model_path, int n_ctx, int threads, int n_gpu_layers, char * err_buf, int err_buf_len) {
    try {
        th_init_backend();

        llama_model_params model_params = llama_model_default_params();
        model_params.n_gpu_layers = n_gpu_layers;
        model_params.use_mmap = true;

        llama_model * model = llama_model_load_from_file(model_path, model_params);
        if (model == nullptr) {
            th_copy_error(err_buf, err_buf_len, "failed to load llama.cpp model");
            return nullptr;
        }

        auto * handle = new th_llama_model();
        handle->model = model;
        handle->n_ctx = n_ctx;
        handle->threads = threads;
        return handle;
    } catch (const std::exception & ex) {
        th_copy_error(err_buf, err_buf_len, ex.what());
        return nullptr;
    }
}

int th_llama_model_predict(void * raw_handle, const char * prompt, int max_tokens, char * out_buf, int out_buf_len, char * err_buf, int err_buf_len) {
    auto * handle = static_cast<th_llama_model *>(raw_handle);
    if (handle == nullptr || handle->model == nullptr) {
        th_copy_error(err_buf, err_buf_len, "llama.cpp model handle is nil");
        return 1;
    }

    llama_context_params ctx_params = llama_context_default_params();
    ctx_params.n_ctx = static_cast<uint32_t>(handle->n_ctx);
    ctx_params.n_batch = static_cast<uint32_t>(std::max(handle->n_ctx, 1));
    ctx_params.n_ubatch = ctx_params.n_batch;
    ctx_params.n_threads = handle->threads;
    ctx_params.n_threads_batch = handle->threads;
    ctx_params.no_perf = true;

    llama_context * ctx = llama_init_from_model(handle->model, ctx_params);
    if (ctx == nullptr) {
        th_copy_error(err_buf, err_buf_len, "failed to create llama.cpp context");
        return 1;
    }

    llama_sampler * smpl = llama_sampler_chain_init(llama_sampler_chain_default_params());
    llama_sampler_chain_add(smpl, llama_sampler_init_greedy());

    const llama_vocab * vocab = llama_model_get_vocab(handle->model);
    const int prompt_len = static_cast<int>(std::strlen(prompt));
    const int n_prompt_tokens = -llama_tokenize(vocab, prompt, prompt_len, nullptr, 0, true, true);
    if (n_prompt_tokens <= 0) {
        llama_sampler_free(smpl);
        llama_free(ctx);
        th_copy_error(err_buf, err_buf_len, "failed to tokenize prompt");
        return 1;
    }

    std::vector<llama_token> prompt_tokens(static_cast<size_t>(n_prompt_tokens));
    if (llama_tokenize(vocab, prompt, prompt_len, prompt_tokens.data(), static_cast<int32_t>(prompt_tokens.size()), true, true) < 0) {
        llama_sampler_free(smpl);
        llama_free(ctx);
        th_copy_error(err_buf, err_buf_len, "failed to tokenize prompt into buffer");
        return 1;
    }

    const int required_ctx = n_prompt_tokens + std::max(max_tokens, 1);
    if (required_ctx > handle->n_ctx) {
        llama_sampler_free(smpl);
        llama_free(ctx);
        th_copy_error(err_buf, err_buf_len, "prompt exceeds local context window; increase local_context_size or reduce prompt size");
        return 1;
    }

    llama_batch batch = llama_batch_get_one(prompt_tokens.data(), static_cast<int32_t>(prompt_tokens.size()));
    std::string response;
    response.reserve(static_cast<size_t>(std::max(max_tokens, 1) * 16));

    llama_token next_token = LLAMA_TOKEN_NULL;
    for (int generated = 0; generated < max_tokens; ++generated) {
        if (llama_decode(ctx, batch) != 0) {
            llama_sampler_free(smpl);
            llama_free(ctx);
            th_copy_error(err_buf, err_buf_len, "llama_decode failed");
            return 1;
        }

        next_token = llama_sampler_sample(smpl, ctx, -1);
        if (llama_vocab_is_eog(vocab, next_token)) {
            break;
        }

        std::vector<char> piece(256);
        int piece_len = llama_token_to_piece(vocab, next_token, piece.data(), static_cast<int32_t>(piece.size()), 0, true);
        if (piece_len < 0) {
            piece.resize(static_cast<size_t>(-piece_len));
            piece_len = llama_token_to_piece(vocab, next_token, piece.data(), static_cast<int32_t>(piece.size()), 0, true);
        }
        if (piece_len < 0) {
            llama_sampler_free(smpl);
            llama_free(ctx);
            th_copy_error(err_buf, err_buf_len, "failed to convert token to piece");
            return 1;
        }

        response.append(piece.data(), static_cast<size_t>(piece_len));
        llama_sampler_accept(smpl, next_token);
        batch = llama_batch_get_one(&next_token, 1);
    }

    llama_sampler_free(smpl);
    llama_free(ctx);

    if (out_buf == nullptr || out_buf_len <= 0) {
        th_copy_error(err_buf, err_buf_len, "invalid output buffer");
        return 1;
    }

    const int copy_len = std::min<int>(static_cast<int>(response.size()), out_buf_len - 1);
    std::memcpy(out_buf, response.data(), static_cast<size_t>(copy_len));
    out_buf[copy_len] = '\0';
    return 0;
}

void th_llama_model_free(void * raw_handle) {
    auto * handle = static_cast<th_llama_model *>(raw_handle);
    if (handle == nullptr) {
        return;
    }

    if (handle->model != nullptr) {
        llama_model_free(handle->model);
        handle->model = nullptr;
    }

    delete handle;
}