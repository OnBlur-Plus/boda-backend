import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { EnvironmentModule } from 'src/module/config/config.module';
import { TypeORMConfigService } from 'src/module/config/typeorm-config.service';
import { HealthController } from 'src/module/health/health.controller';
import { StreamModule } from 'src/module/stream/stream.module';

@Module({
  imports: [
    EnvironmentModule,
    TypeOrmModule.forRootAsync({
      useExisting: TypeORMConfigService,
    }),
    StreamModule,
  ],
  controllers: [HealthController],
  providers: [],
  exports: [],
})
export class MainModule {}