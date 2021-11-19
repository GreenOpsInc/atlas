package com.greenops.util.error;

public class AtlasAuthenticationError extends RuntimeException {

    public AtlasAuthenticationError() {
        super();
    }

    public AtlasAuthenticationError(String message) {
        super(message);
    }

    public AtlasAuthenticationError(Throwable throwable) {
        super(throwable);
    }

    public AtlasAuthenticationError(String message, Throwable throwable) {
        super(message, throwable);
    }
}
