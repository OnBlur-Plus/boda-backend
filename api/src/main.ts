import { ClassSerializerInterceptor, HttpStatus, UnprocessableEntityException, ValidationPipe } from '@nestjs/common';
import { NestFactory, Reflector } from '@nestjs/core';
import { ExpressAdapter, NestExpressApplication } from '@nestjs/platform-express';
import { EnvConfigService } from 'src/module/config/env-config.service';
import { MainModule } from 'src/module/main.module';

async function bootstrap() {
  const app = await NestFactory.create<NestExpressApplication>(MainModule, new ExpressAdapter(), { cors: true });

  const environmentService = app.get(EnvConfigService);

  const port = environmentService.lport;

  // app.use(helmet({ contentSecurityPolicy: false }));

  const reflector = app.get(Reflector);

  app.useGlobalInterceptors(new ClassSerializerInterceptor(reflector));

  app.useGlobalPipes(
    new ValidationPipe({
      whitelist: true,
      errorHttpStatusCode: HttpStatus.UNPROCESSABLE_ENTITY,
      transform: true,
      dismissDefaultMessages: true,
      forbidNonWhitelisted: true,
      forbidUnknownValues: true,
      disableErrorMessages: false,
      validationError: {
        target: true,
        value: true,
      },
      exceptionFactory: (errors) => new UnprocessableEntityException(errors),
    }),
  );

  await app.listen(port);
}

void bootstrap();
