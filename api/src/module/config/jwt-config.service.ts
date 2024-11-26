import { Injectable } from '@nestjs/common';
import { AbstractConfigService } from '@nestjs-library/config';
import { Expose, Transform } from 'class-transformer';
import { IsNotEmpty, IsString } from 'class-validator';

@Injectable()
export class JwtConfigService extends AbstractConfigService<JwtConfigService> {
  @Expose({ name: 'JWT_SECRET_KEY' })
  @Transform(({ value }) => value ?? '')
  @IsString()
  @IsNotEmpty()
  secretKey: string;

  @Expose({ name: 'JWT_EXPIRES_IN' })
  @Transform(({ value }) => value ?? '1h')
  @IsString()
  @IsNotEmpty()
  expiresIn: string;
}
