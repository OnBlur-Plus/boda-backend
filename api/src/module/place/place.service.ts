import { Injectable } from '@nestjs/common';
import { CreatePlaceDto } from './dto/create-place.dto';
import { UpdatePlaceDto } from './dto/update-place.dto';
import { Repository } from 'typeorm';
import { Place } from 'src/module/place/entities/place.entity';
import { InjectRepository } from '@nestjs/typeorm';

@Injectable()
export class PlaceService {
  constructor(@InjectRepository(Place) private readonly placeRepository: Repository<Place>) {}

  async create(createPlaceDto: CreatePlaceDto) {
    return await this.placeRepository.create(createPlaceDto);
  }

  findAll() {
    return `This action returns all place`;
  }

  findOne(id: number) {
    return `This action returns a #${id} place`;
  }

  remove(id: number) {
    return `This action removes a #${id} place`;
  }

  async update(id: number, updatePlaceDto: UpdatePlaceDto) {
    await this.placeRepository.findOneByOrFail({ id });

    return await this.placeRepository.update({ id }, updatePlaceDto);
  }
}
