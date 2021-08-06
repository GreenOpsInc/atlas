package com.greenops.util.error;

public class AtlasNonRetryableError extends RuntimeException {

    public AtlasNonRetryableError(String message) {
        super(message);
    }

    public AtlasNonRetryableError(Throwable throwable) {
        super(throwable);
    }

    public AtlasNonRetryableError(String message, Throwable throwable) {
        super(message, throwable);
    }
}
