import { IsNumber } from 'class-validator';
export class EndAccidentDto {
  @IsNumber()
  id: number;
}
