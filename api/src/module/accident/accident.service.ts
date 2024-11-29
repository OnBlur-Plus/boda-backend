import { Injectable } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Accident } from './entities/accident.entities';
import { Between, Repository } from 'typeorm';

@Injectable()
export class AccidentService {
  constructor(@InjectRepository(Accident) private readonly accidentRepository: Repository<Accident>) {}

  async getAccidents(pageNum: number, pageSize: number) {
    return await this.accidentRepository.find({
      skip: (pageNum - 1) * pageSize,
      take: pageSize,
      order: { startAt: 'DESC' },
    });
  }

  async getAccidentsByStreamKey(streamKey: string) {
    return await this.accidentRepository.find({ where: { stream: { streamKey } } });
  }

  async getAccidentsByDate(date: Date) {
    const startDate = new Date(date.setUTCHours(0, 0, 0));
    const endDate = new Date(date.setUTCHours(23, 59, 59));

    return await this.accidentRepository.find({
      where: { startAt: Between(startDate, endDate) },
      order: { startAt: 'DESC' },
    });
  }

  async getAccidentWithStream(id: number) {
    return await this.accidentRepository.findOne({ where: { id }, relations: ['stream'] });
  }

  async createAccident(accident: Omit<Accident, 'id' | 'createdAt' | 'updatedAt'>) {
    return await this.accidentRepository.save(accident);
  }
}
