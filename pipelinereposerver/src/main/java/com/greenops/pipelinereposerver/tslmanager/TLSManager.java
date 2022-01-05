package com.greenops.pipelinereposerver.tslmanager;

public interface TLSManager {
    // TODO: find *x509.CertPool analogue in java
    public /*x509.CertPool*/ void bestEffortSystemCertPool();

    // TODO: find tls.Config analogue in java
    public /*tls.Config*/ void getServerTLSConf();

    // TODO: find a way to put function or something else as a callback
    public void watchServerTLSConf(String handler /* func(conf *tls.Config, err error) */);
}
