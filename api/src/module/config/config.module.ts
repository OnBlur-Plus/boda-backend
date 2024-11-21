import { Global, Module } from '@nestjs/common';
import { ConfigModule, DotenvSourceProvider } from '@nestjs-library/config';

import { NodeEnv } from 'src/constant/node-env.enum';
import path from 'path';
import { DatabaseConfigService } from 'src/module/config/database-config.service';
import { EnvConfigService } from 'src/module/config/env-config.service';
import { JwtConfigService } from 'src/module/config/jwt-config.service';

const getEnvFilePath = () =>
  path.join(
    process.cwd(),
    `.env.${
      (process.env.NODE_ENV === NodeEnv.LOCAL
        ? NodeEnv.DEVELOPMENT
        : process.env.NODE_ENV) ?? NodeEnv.LOCAL
    }`,
  );

@Global()
@Module({
  imports: [
    ConfigModule.forFeature(
      [DatabaseConfigService, EnvConfigService, JwtConfigService],
      {
        // envFilePath: [getEnvFilePath()],
        global: true,
        sourceProvider: new DotenvSourceProvider({ path: getEnvFilePath() }),
      },
    ),
  ],
  providers: [DatabaseConfigService, EnvConfigService, JwtConfigService],
  exports: [DatabaseConfigService, EnvConfigService, JwtConfigService],
})
export class EnvironmentModule {}
