// database-config.service.ts
import { Injectable } from '@nestjs/common';
import { AbstractConfigService } from '@nestjs-library/config';
import { Expose, Transform, Type } from 'class-transformer';
import { IsNotEmpty, IsPositive, IsString } from 'class-validator';

@Injectable()
export class DatabaseConfigService extends AbstractConfigService<DatabaseConfigService> {
  @Expose({ name: 'DATABASE_HOST' })
  @Transform(({ value }) => value ?? 'localhost')
  @IsString()
  @IsNotEmpty()
  host: string;

  @Expose({ name: 'DATABASE_PORT' })
  @Type(() => Number)
  @Transform(({ value }) => value ?? 5532)
  @IsPositive()
  port: number;

  @Expose({ name: 'DATABASE_PASSWORD' })
  @Transform(({ value }) => value ?? 'local')
  @IsString()
  @IsNotEmpty()
  password: string;
}
