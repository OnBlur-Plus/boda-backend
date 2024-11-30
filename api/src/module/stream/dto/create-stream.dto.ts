import { IsEnum, IsOptional, IsString } from 'class-validator';
import { StreamStatus } from '../entities/stream.entity';

export class CreateStreamDto {
  @IsString()
  title: string;

  @IsString()
  subTitle: string;

  @IsString()
  @IsOptional()
  thumbnailUrl: string;
}
