import { IsString, Matches } from 'class-validator';

export class VerifyDto {
  @IsString()
  @Matches(/^[0-9]{4}$/)
  pin: string;
}
