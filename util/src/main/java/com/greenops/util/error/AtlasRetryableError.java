package com.greenops.util.error;

public class AtlasRetryableError extends RuntimeException {

    public AtlasRetryableError(String message) {
        super(message);
    }

    public AtlasRetryableError(Throwable throwable) {
        super(throwable);
    }

    public AtlasRetryableError(String message, Throwable throwable) {
        super(message, throwable);
    }
}
