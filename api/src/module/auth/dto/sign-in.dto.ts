import { IsString, Matches } from 'class-validator';

export class SignInDto {
  @IsString()
  @Matches(/^[0-9]{4}$/)
  pin: string;
}
