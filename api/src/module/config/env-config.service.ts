// database-config.service.ts
import { Injectable } from '@nestjs/common';
import { AbstractConfigService } from '@nestjs-library/config';
import { Expose, Transform, Type } from 'class-transformer';
import { IsPositive } from 'class-validator';

@Injectable()
export class EnvConfigService extends AbstractConfigService<EnvConfigService> {
  @Expose({ name: 'LPORT' })
  @Type(() => Number)
  @Transform(({ value }) => value ?? 3000)
  @IsPositive()
  lport: number;
}
