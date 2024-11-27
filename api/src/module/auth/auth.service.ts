import { Repository } from 'typeorm';
import { Injectable, UnauthorizedException } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { JwtService } from '@nestjs/jwt';
import { User } from './entities/user.entity';
import { VerifyDto } from './dto/verify.dto';

@Injectable()
export class AuthService {
  constructor(
    private readonly jwtService: JwtService,
    @InjectRepository(User) private readonly userRepository: Repository<User>,
  ) {}

  async verify({ pin }: VerifyDto) {
    const user = await this.userRepository.findOne({ where: { pin } });

    if (!user) {
      throw new UnauthorizedException();
    }

    return { access_token: await this.jwtService.signAsync({ name: user.name }) };
  }
}
