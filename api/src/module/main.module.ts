import { Module } from '@nestjs/common';
import { EnvironmentModule } from 'src/module/config/config.module';
import { HealthController } from 'src/module/health/health.controller';

@Module({
  imports: [EnvironmentModule],
  controllers: [HealthController],
  providers: [],
  exports: [],
})
export class MainModule {}
