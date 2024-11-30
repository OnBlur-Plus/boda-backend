import { IsEnum, IsOptional, IsString } from 'class-validator';
import { AccidentLevel, AccidentType } from '../entities/accident.entities';

export class UpdateAccidentDto {
  @IsEnum(AccidentType)
  @IsOptional()
  type: AccidentType;

  @IsEnum(AccidentLevel)
  @IsOptional()
  level: AccidentLevel;

  @IsString()
  @IsOptional()
  reason: string;
}
