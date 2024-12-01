import { IsNumber, IsString } from 'class-validator';
export class EndAccidentDto {
  @IsNumber()
  id: number;

  @IsString()
  videoUrl: string;
}
