import { Injectable } from '@nestjs/common';
import { AbstractConfigService, OptionalBoolean } from '@nestjs-library/config';
import { Expose, Transform, Type } from 'class-transformer';
import { IsBoolean, IsNotEmpty, IsPositive, IsString } from 'class-validator';
import { MysqlConnectionOptions } from 'typeorm/driver/mysql/MysqlConnectionOptions';

@Injectable()
export class TypeORMConfigService
  extends AbstractConfigService<TypeORMConfigService>
  implements MysqlConnectionOptions
{
  type = 'mariadb' as const;

  @Expose({ name: 'DATABASE_HOST' })
  @Transform(({ value }) => value ?? 'localhost')
  @IsString()
  @IsNotEmpty()
  host: string;

  @Expose({ name: 'DATABASE_PORT' })
  @Type(() => Number)
  @Transform(({ value }) => value ?? 3306)
  @IsPositive()
  port: number;

  @Expose({ name: 'DATABASE_USERNAME' })
  @Transform(({ value }) => value ?? 'root')
  @IsString()
  @IsNotEmpty()
  username: string;

  @Expose({ name: 'DATABASE_PASSWORD' })
  @Transform(({ value }) => value ?? 'local')
  @IsString()
  @IsNotEmpty()
  password: string;

  @Expose({ name: 'DATABASE_NAME' })
  @Transform(({ value }) => value ?? 'boda')
  @IsString()
  @IsNotEmpty()
  database: string;

  @Expose({ name: 'DATABASE_AUTO_LOAD_ENTITIES' })
  @Transform(({ value }) => OptionalBoolean(value) ?? true)
  @IsBoolean()
  @IsNotEmpty()
  autoLoadEntities: boolean;

  @Expose({ name: 'DATABASE_SYNCHRONIZE' })
  @Transform(({ value }) => OptionalBoolean(value) ?? false)
  @IsBoolean()
  @IsNotEmpty()
  synchronize: boolean;
}
