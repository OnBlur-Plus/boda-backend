import { IsOptional, IsString } from 'class-validator';

export class UpdateStreamDto {
  @IsString()
  @IsOptional()
  title: string;

  @IsString()
  @IsOptional()
  subTitle: string;
}
