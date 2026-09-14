import assert from "node:assert/strict";
import test from "node:test";

import { generationErrorMessage } from "../src/lib/generation-error.ts";

test("上游 5xx 提示指向渠道与模型，不再笼统报网络异常", () => {
    const message = generationErrorMessage(new Error("模型服务暂时不可用（HTTP 502）；上游：上游 HTTP 502"));

    assert.match(message, /HTTP 502/);
    assert.match(message, /渠道/);
    assert.ok(!message.includes("网络异常"), `不应再归为网络异常：${message}`);
});

test("503 与 504 同样给出带状态码的服务端提示", () => {
    assert.match(generationErrorMessage(new Error("模型服务暂时不可用（HTTP 503）；上游：x")), /HTTP 503/);
    assert.match(generationErrorMessage(new Error("模型服务暂时不可用（HTTP 504）；上游：x")), /HTTP 504/);
});

test("真实网络故障仍报网络异常", () => {
    assert.equal(generationErrorMessage(new Error("dial tcp 127.0.0.1:8080: connect: connection refused")), "网络异常。");
    assert.equal(generationErrorMessage(new Error("fetch failed")), "网络异常。");
});

test("限流与鉴权提示不受影响", () => {
    assert.equal(generationErrorMessage(new Error("Request failed with status code 429")), "服务当前繁忙，请稍后重试。");
    assert.equal(generationErrorMessage(new Error("Request failed with status code 401")), "生成服务鉴权失败，请检查渠道配置。");
    assert.equal(generationErrorMessage(new Error("Request failed with status code 404")), "生成服务地址不可用，请检查渠道配置。");
});
