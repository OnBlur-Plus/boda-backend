import { Injectable, UnauthorizedException } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { User } from './entities/user.entity';
import { Repository } from 'typeorm';
import { JwtService } from '@nestjs/jwt';

@Injectable()
export class AuthService {
  constructor(
    private readonly jwtService: JwtService,
    @InjectRepository(User) private readonly userRepository: Repository<User>,
  ) {}

  async signIn(pin: string) {
    const user = await this.userRepository.findOne({ where: { pin } });

    if (!user) {
      throw new UnauthorizedException();
    }

    return { access_token: await this.jwtService.signAsync({ name: user.name }) };
  }
}
