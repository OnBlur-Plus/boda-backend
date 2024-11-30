import { Global, Module } from '@nestjs/common';
import { ConfigModule, DotenvSourceProvider } from '@nestjs-library/config';

import path from 'path';
import { TypeORMConfigService } from 'src/module/config/typeorm-config.service';
import { EnvConfigService } from 'src/module/config/env-config.service';
import { JwtConfigService } from 'src/module/config/jwt-config.service';
import { FirebaseConfigService } from './firebase-config.service';

export const getEnvFilePath = () => path.join(process.cwd(), `.env`);

@Global()
@Module({
  imports: [
    ConfigModule.forFeature([TypeORMConfigService, EnvConfigService, JwtConfigService, FirebaseConfigService], {
      global: true,
      sourceProvider: new DotenvSourceProvider({ path: getEnvFilePath() }),
    }),
  ],
})
export class EnvironmentModule {}
