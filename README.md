# Translate-Agent

## TypeScript client

### Installing

#### npm
```bash
npm config set @buf:registry  https://buf.build/gen/npm/v1/

npm install @buf/expectdigital_translate-agent.bufbuild_connect-es@latest
npm install @buf/expectdigital_translate-agent.bufbuild_es@latest
```
#### Other

Other package managers and their instructions can be found at [buf.build registry](https://buf.build/expectdigital/translate-agent/assets/main) under the `Node.js` tab.
Packages that are needed:
- `bufbuild/connect-es`
- `bufbuild/es`

### Usage of TypeScript client
```typescript
import { createPromiseClient } from '@bufbuild/connect';
import { createGrpcWebTransport } from '@bufbuild/connect-web';

import { TranslateService } from '@buf/expectdigital_translate-agent.bufbuild_connect-es/translate/v1/translate_connect';
import { CreateServiceRequest, Service } from '@buf/expectdigital_translate-agent.bufbuild_es/translate/v1/translate_pb';


const transport = createGrpcWebTransport({
  baseUrl: "http://localhost:8080",
});

const client = createPromiseClient(TranslateService, transport);

client.createService({
  service: new Service({ name: "test1" })
})
```
