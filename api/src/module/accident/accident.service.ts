import { Injectable, NotFoundException } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Accident } from './entities/accident.entity';
import { Between, Repository } from 'typeorm';
import { UpdateAccidentDto } from './dto/update-accident.dto';
import { StartAccidentDto } from 'src/module/accident/dto/start-accident.dto';
import { StreamService } from 'src/module/stream/stream.service';

@Injectable()
export class AccidentService {
  constructor(
    @InjectRepository(Accident) private readonly accidentRepository: Repository<Accident>,
    private readonly streamService: StreamService,
  ) {}

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

  async startAccident(startAccidentDto: StartAccidentDto) {
    const { type, reason, streamKey } = startAccidentDto;
    const stream = this.streamService.findStream(streamKey);

    if (!stream) {
      throw new NotFoundException('Not Found Stream');
    }

    const accident = this.accidentRepository.create({
      type,
      reason,
      startAt: new Date(),
      level: 1,
      streamKey,
    });

    return await this.accidentRepository.save(accident);
  }
  async endAccident(id: number) {
    return await this.accidentRepository.update(id, { endAt: Date() });
  }

  async updateAccident(accidentId: number, updateaccidentDto: UpdateAccidentDto) {
    return await this.accidentRepository.update(accidentId, updateaccidentDto);
  }

  async deleteAccident(accidentId: number) {
    return await this.accidentRepository.delete(accidentId);
  }
}
