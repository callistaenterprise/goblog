package model

type DiscoveryToken struct {
        State   string `json:"state"` // UP, DOWN ??
        Address string `json:"address"`
}

type Instance struct {
        Name string
        breakers []CircuitBreaker
        pools []ThreadPool
}

type HystrixProducer struct {
        Ip string
        State string
}

type ThreadPool struct {
        CurrentCorePoolSize                                        int    `json:"currentCorePoolSize"`
        CurrentLargestPoolSize                                     int    `json:"currentLargestPoolSize"`
        CurrentActiveCount                                         int    `json:"currentActiveCount"`
        PropertyValueMetricsRollingStatisticalWindowInMilliseconds string `json:"propertyValue_metricsRollingStatisticalWindowInMilliseconds"`
        CurrentMaximumPoolSize                                     int    `json:"currentMaximumPoolSize"`
        CurrentQueueSize                                           int    `json:"currentQueueSize"`
        Type                                                       string `json:"type"`
        CurrentTaskCount                                           int    `json:"currentTaskCount"`
        TypeAndName                                                string `json:"TypeAndName"`
        CurrentCompletedTaskCount                                  int    `json:"currentCompletedTaskCount"`
        RollingMaxActiveThreads                                    int    `json:"rollingMaxActiveThreads"`
        InstanceID                                                 string `json:"instanceId"`
        InstanceKey                                                string `json:"InstanceKey"`
        Name                                                       string `json:"name"`
        ReportingHosts                                             int    `json:"reportingHosts"`
        CurrentPoolSize                                            int    `json:"currentPoolSize"`
        PropertyValueQueueSizeRejectionThreshold                   string `json:"propertyValue_queueSizeRejectionThreshold"`
        RollingCountThreadsExecuted                                int    `json:"rollingCountThreadsExecuted"`
}

type CircuitBreaker struct {
        RollingCountFallbackFailure                                int    `json:"rollingCountFallbackFailure"`
        RollingCountFallbackSuccess                                int    `json:"rollingCountFallbackSuccess"`
        PropertyValueCircuitBreakerRequestVolumeThreshold          string `json:"propertyValue_circuitBreakerRequestVolumeThreshold"`
        PropertyValueCircuitBreakerForceOpen                       bool   `json:"propertyValue_circuitBreakerForceOpen"`
        PropertyValueMetricsRollingStatisticalWindowInMilliseconds string `json:"propertyValue_metricsRollingStatisticalWindowInMilliseconds"`
        LatencyTotalMean                                           int    `json:"latencyTotal_mean"`
        Type                                                       string `json:"type"`
        RollingCountResponsesFromCache                             int    `json:"rollingCountResponsesFromCache"`
        TypeAndName                                                string `json:"TypeAndName"`
        RollingCountTimeout                                        int    `json:"rollingCountTimeout"`
        PropertyValueExecutionIsolationStrategy                    string `json:"propertyValue_executionIsolationStrategy"`
        InstanceID                                                 string `json:"instanceId"`
        RollingCountFailure                                        int    `json:"rollingCountFailure"`
        RollingCountExceptionsThrown                               int    `json:"rollingCountExceptionsThrown"`
        LatencyExecuteMean                                         int    `json:"latencyExecute_mean"`
        IsCircuitBreakerOpen                                       bool   `json:"isCircuitBreakerOpen"`
        ErrorCount                                                 int    `json:"errorCount"`
        Group                                                      string `json:"group"`
        RollingCountSemaphoreRejected                              int    `json:"rollingCountSemaphoreRejected"`
        LatencyTotal                                               struct {
                                                                           Num0   int `json:"0"`
                                                                           Num25  int `json:"25"`
                                                                           Num50  int `json:"50"`
                                                                           Num75  int `json:"75"`
                                                                           Num90  int `json:"90"`
                                                                           Num95  int `json:"95"`
                                                                           Num99  int `json:"99"`
                                                                           Num100 int `json:"100"`
                                                                           Nine95 int `json:"99.5"`
                                                                   } `json:"latencyTotal"`
        RequestCount                  int `json:"requestCount"`
        RollingCountCollapsedRequests int `json:"rollingCountCollapsedRequests"`
        RollingCountShortCircuited    int `json:"rollingCountShortCircuited"`
        LatencyExecute                struct {
                                                                           Num0   int `json:"0"`
                                                                           Num25  int `json:"25"`
                                                                           Num50  int `json:"50"`
                                                                           Num75  int `json:"75"`
                                                                           Num90  int `json:"90"`
                                                                           Num95  int `json:"95"`
                                                                           Num99  int `json:"99"`
                                                                           Num100 int `json:"100"`
                                                                           Nine95 int `json:"99.5"`
                                                                   } `json:"latencyExecute"`
        PropertyValueCircuitBreakerSleepWindowInMilliseconds          string `json:"propertyValue_circuitBreakerSleepWindowInMilliseconds"`
        CurrentConcurrentExecutionCount                               int    `json:"currentConcurrentExecutionCount"`
        PropertyValueExecutionIsolationSemaphoreMaxConcurrentRequests string `json:"propertyValue_executionIsolationSemaphoreMaxConcurrentRequests"`
        ErrorPercentage                                               int    `json:"errorPercentage"`
        RollingCountThreadPoolRejected                                int    `json:"rollingCountThreadPoolRejected"`
        PropertyValueCircuitBreakerEnabled                            bool   `json:"propertyValue_circuitBreakerEnabled"`
        PropertyValueExecutionIsolationThreadInterruptOnTimeout       bool   `json:"propertyValue_executionIsolationThreadInterruptOnTimeout"`
        PropertyValueRequestCacheEnabled                              bool   `json:"propertyValue_requestCacheEnabled"`
        RollingCountFallbackRejection                                 int    `json:"rollingCountFallbackRejection"`
        PropertyValueRequestLogEnabled                                bool   `json:"propertyValue_requestLogEnabled"`
        RollingCountSuccess                                           int    `json:"rollingCountSuccess"`
        PropertyValueFallbackIsolationSemaphoreMaxConcurrentRequests  string `json:"propertyValue_fallbackIsolationSemaphoreMaxConcurrentRequests"`
        InstanceKey                                                   string `json:"InstanceKey"`
        PropertyValueCircuitBreakerErrorThresholdPercentage           string `json:"propertyValue_circuitBreakerErrorThresholdPercentage"`
        PropertyValueCircuitBreakerForceClosed                        bool   `json:"propertyValue_circuitBreakerForceClosed"`
        Name                                                          string `json:"name"`
        ReportingHosts                                                int    `json:"reportingHosts"`
        PropertyValueExecutionIsolationThreadPoolKeyOverride          string `json:"propertyValue_executionIsolationThreadPoolKeyOverride"`
        PropertyValueExecutionIsolationThreadTimeoutInMilliseconds    string `json:"propertyValue_executionIsolationThreadTimeoutInMilliseconds"`
}
