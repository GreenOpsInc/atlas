package com.greenops.util.error;

public class AtlasBadKeyError extends RuntimeException {

    public AtlasBadKeyError() {
        super();
    }

    public AtlasBadKeyError(String message) {
        super(message);
    }

    public AtlasBadKeyError(Throwable throwable) {
        super(throwable);
    }

    public AtlasBadKeyError(String message, Throwable throwable) {
        super(message, throwable);
    }
}
