import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { EnvironmentModule } from 'src/module/config/config.module';
import { TypeORMConfigService } from 'src/module/config/typeorm-config.service';
import { AuthGuard } from './auth/auth.guard';
import { AccidentModule } from './accident/accident.module';
import { HealthModule } from './health/health.module';
import { StreamModule } from 'src/module/stream/stream.module';
import { AuthModule } from './auth/auth.module';
import { NotificationModule } from './notification/notification.module';

@Module({
  imports: [
    EnvironmentModule,
    TypeOrmModule.forRootAsync({
      useExisting: TypeORMConfigService,
    }),
    AccidentModule,
    HealthModule,
    StreamModule,
    AuthModule,
    NotificationModule,
  ],
  providers: [{ provide: 'APP_GUARD', useClass: AuthGuard }],
  exports: [],
})
export class MainModule {}
