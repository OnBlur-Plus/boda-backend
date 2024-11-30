import { IsEnum, IsOptional, IsString, IsUUID } from 'class-validator';
import { StreamStatus } from '../entities/stream.entity';

export class CreateStreamDto {
  @IsUUID()
  streamKey: string;

  @IsString()
  title: string;

  @IsString()
  @IsOptional()
  subTitle: string;

  @IsString()
  @IsOptional()
  thumbnailUrl: string;

  @IsEnum(StreamStatus)
  status: StreamStatus;
}
