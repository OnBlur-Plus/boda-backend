import { Injectable } from '@nestjs/common';
import { AbstractConfigService } from '@nestjs-library/config';
import { Expose, Transform } from 'class-transformer';
import { IsNotEmpty, IsString } from 'class-validator';

@Injectable()
export class FirebaseConfigService extends AbstractConfigService<FirebaseConfigService> {
  @Expose({ name: 'FIREBASE_PROJECT_ID' })
  @Transform(({ value }) => value ?? '')
  @IsString()
  @IsNotEmpty()
  projectId: string;

  @Expose({ name: 'FIREBASE_CLIENT_EMAIL' })
  @Transform(({ value }) => value ?? '')
  @IsString()
  @IsNotEmpty()
  clientEmail: string;

  @Expose({ name: 'FIREBASE_PRIVATE_KEY' })
  @Transform(({ value }) => value ?? '')
  @IsString()
  @IsNotEmpty()
  privateKey: string;
}
