import { IsString } from 'class-validator';

export class CreatePlaceDto {
  @IsString()
  name: string;
}
