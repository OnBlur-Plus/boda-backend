import { DynamicModule, Global, Module } from '@nestjs/common';
import { ConfigModule, DotenvSourceProvider } from '@nestjs-library/config';

import { NodeEnv } from 'src/constant/node-env.enum';
import path from 'path';
import { TypeORMConfigService } from 'src/module/config/typeorm-config.service';
import { EnvConfigService } from 'src/module/config/env-config.service';
import { JwtConfigService } from 'src/module/config/jwt-config.service';

const getEnvFilePath = () =>
  path.join(
    process.cwd(),
    `.env.${(process.env.NODE_ENV === NodeEnv.LOCAL ? NodeEnv.DEVELOPMENT : process.env.NODE_ENV) ?? NodeEnv.LOCAL}`,
  );

@Global()
@Module({})
export class EnvironmentModule {
  static forRoot(): DynamicModule {
    return {
      module: EnvironmentModule,
      imports: [
        ConfigModule.forFeature([TypeORMConfigService, EnvConfigService, JwtConfigService], {
          global: true,
          sourceProvider: new DotenvSourceProvider({ path: getEnvFilePath() }),
        }),
      ],
      providers: [TypeORMConfigService, EnvConfigService, JwtConfigService],
      exports: [TypeORMConfigService, EnvConfigService, JwtConfigService],
    };
  }
}
