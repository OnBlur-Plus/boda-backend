import { IsEnum, IsOptional, IsString } from 'class-validator';
import { AccidentLevel, AccidentType } from '../entities/accident.entity';

export class UpdateAccidentDto {
  @IsEnum(AccidentType)
  @IsOptional()
  type: AccidentType;

  @IsEnum(AccidentLevel)
  @IsOptional()
  level: AccidentLevel;

  @IsString()
  reason: string;
}
