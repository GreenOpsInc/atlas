package com.greenops.workfloworchestrator.ingest.apiclient.util;

import com.greenops.workfloworchestrator.error.AtlasNonRetryableError;
import com.greenops.workfloworchestrator.error.AtlasRetryableError;
import org.apache.http.HttpResponse;

public class ApiClientUtil {

    public static void checkResponseStatus(HttpResponse httpResponse) {
        switch (httpResponse.getStatusLine().getStatusCode()) {
            case 400:
                throw new AtlasNonRetryableError("Returned with bad request");
            case 404:
                throw new AtlasNonRetryableError("Returned with not found");
            case 500:
                throw new AtlasRetryableError("Returned internal service error");
            case 503:
                throw new AtlasRetryableError("Service is down");
        }
    }
}
