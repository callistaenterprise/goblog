package se.callista.microservises.support.edge;

import javax.net.ssl.HttpsURLConnection;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.boot.builder.SpringApplicationBuilder;
import org.springframework.cloud.netflix.zuul.EnableZuulProxy;
import org.springframework.stereotype.Controller;

@SpringBootApplication
@Controller
@EnableZuulProxy
// @EnableResourceServer
public class ZuulApplication {

    private static final Logger LOG = LoggerFactory.getLogger(ZuulApplication.class);

    static {
        // for localhost testing only
        LOG.warn("Will now disable hostname check in SSL, only to be used during development");
        HttpsURLConnection.setDefaultHostnameVerifier((hostname, sslSession) -> true);
    }

    public static void main(String[] args) {
        int buildNo = 15;
        LOG.info("Edge-server, starting build no. {}...", buildNo);

        new SpringApplicationBuilder(ZuulApplication.class).web(true).run(args);

        LOG.info("Edge-server, build no. {} started HEJ HEJ HEJ", buildNo);
    }

//    @Bean
//    public Sampler defaultSampler() {
//        return new AlwaysSampler();
//    }
}
