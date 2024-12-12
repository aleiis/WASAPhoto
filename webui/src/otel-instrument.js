import { WebTracerProvider } from '@opentelemetry/sdk-trace-web';
import { ConsoleSpanExporter, SimpleSpanProcessor } from '@opentelemetry/sdk-trace-base';
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-http';
import { FetchInstrumentation } from '@opentelemetry/instrumentation-fetch';
import { XMLHttpRequestInstrumentation } from '@opentelemetry/instrumentation-xml-http-request';
import { registerInstrumentations } from '@opentelemetry/instrumentation';
import { Resource } from '@opentelemetry/resources';
import { ATTR_SERVICE_NAME, ATTR_SERVICE_VERSION } from '@opentelemetry/semantic-conventions';
import { W3CTraceContextPropagator } from '@opentelemetry/core';

const OTLPtraceExporter = new OTLPTraceExporter({
  url: 'http://localhost:4318/v1/traces',
});

const traceExporter = new ConsoleSpanExporter();

const resource = Resource.default().merge(
  new Resource({
    [ATTR_SERVICE_NAME]: 'frontend-webui',
    [ATTR_SERVICE_VERSION]: '0.1.0',
  }),
);

const provider = new WebTracerProvider({
  resource: resource,
  spanProcessors: [
    new SimpleSpanProcessor(traceExporter),
    new SimpleSpanProcessor(OTLPtraceExporter) 
  ],
});

provider.register({
  propagator: new W3CTraceContextPropagator(),
});

registerInstrumentations({
  instrumentations: [
    new FetchInstrumentation(),
    new XMLHttpRequestInstrumentation({
      propagateTraceHeaderCorsUrls: /.*/,
    }),
  ],
});