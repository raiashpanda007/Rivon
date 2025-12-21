import { type Request, type Response, type NextFunction, text } from "express";
import { z as zod } from "zod";
import nodemailer from "nodemailer";
import dotenv from "dotenv";
dotenv.config();


const USER_EMAIL = process.env.USER_EMAIL ?? "your@gmail.com";
const EMAIL_APP_PASSWORD = process.env.EMAIL_APP_PASSWORD ?? "email-password";
const MailHandlerTypes = zod.object({
  email: zod.email(),
  subject: zod.string(),
  body: zod.string(),
})

const transporter = nodemailer.createTransport({
  service: 'gmail',
  auth: {
    user: USER_EMAIL,
    pass: EMAIL_APP_PASSWORD
  }
});

export default function MailHandler(req: Request, res: Response, next: NextFunction) {
  const parsedBody = MailHandlerTypes.safeParse(req.body);
  console.log(req.body);
  if (!parsedBody.success) {
    return res.status(422).json({ Message: "Please provide a valid data", Data: parsedBody.error })
  }
  const { body, subject, email } = parsedBody.data;

  const mailOptions = {
    from: USER_EMAIL,
    to: email,
    subject: subject,
    text: body
  };
  transporter.sendMail(mailOptions, function(error, info) {
    if (error) {
      console.log(error);
      return res.status(500).json({
        Message: "Internal Server Error ...",
        Data: error
      })
    } else {
      console.log('Email sent: ' + info.response);
      return res.status(200).json({
        Message: "Mail Sent ... "
      })
    }
  });
}

