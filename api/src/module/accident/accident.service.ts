import { Injectable } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Accident } from './entities/accident.entities';
import { Repository } from 'typeorm';

@Injectable()
export class AccidentService {
  constructor(@InjectRepository(Accident) private readonly accidentRepository: Repository<Accident>) {}
}
