package services

import (
	"fmt"

	"github.com/prefeitura-rio/app-go-api/internal/models"
)

// EmailTemplate defines the structure for email content
type EmailTemplate struct {
	Subject string
	Body    string
	IsHTML  bool
}

// GetEnrollmentPendingEmailTemplate returns email template for "Em Análise" status
func GetEnrollmentPendingEmailTemplate(inscricao *models.Inscricao, curso *models.Curso) EmailTemplate {
	subject := fmt.Sprintf("Inscrição recebida! - %s", curso.Titulo)

	body := fmt.Sprintf(`Olá, %s!

Sua inscrição foi recebida com sucesso e agora está em análise.

Estamos avaliando as solicitações e entraremos em contato em breve para informar o resultado. Você pode acompanhar o status da sua inscrição a qualquer momento na seção "Meus Cursos" na plataforma do Ciclo Carioca.

Até logo,
Equipe do Ciclo Carioca`,
		inscricao.Name,
	)

	return EmailTemplate{
		Subject: subject,
		Body:    body,
		IsHTML:  false,
	}
}

// GetEnrollmentApprovedEmailTemplate returns email template for "Inscrito" status
func GetEnrollmentApprovedEmailTemplate(inscricao *models.Inscricao, curso *models.Curso) EmailTemplate {
	subject := fmt.Sprintf("Parabéns! Sua inscrição foi aprovada - %s", curso.Titulo)

	body := fmt.Sprintf(`Olá, %s!

Temos uma ótima notícia! Sua inscrição foi aprovada.

Você está confirmado(a) na atividade. Mantenha-se atento(a) ao seu e-mail e/ou telefone, caso informações adicionais sejam necessárias para a atividade.

Estamos felizes por ter você conosco!

Até logo,
Equipe do Ciclo Carioca`,
		inscricao.Name,
	)

	return EmailTemplate{
		Subject: subject,
		Body:    body,
		IsHTML:  false,
	}
}

// GetEnrollmentRejectedEmailTemplate returns email template for "Recusado" status
func GetEnrollmentRejectedEmailTemplate(inscricao *models.Inscricao, curso *models.Curso) EmailTemplate {
	subject := fmt.Sprintf("Informações sobre sua inscrição - %s", curso.Titulo)

	body := fmt.Sprintf(`Olá, %s!

Agradecemos o seu interesse na atividade "%s".

Analisamos sua inscrição, mas, infelizmente, ela não foi aprovada desta vez. A recusa pode ter ocorrido por não ter sido cumprido algum dos requisitos ou critérios definidos pela unidade responsável.

Não desanime! Convidamos você a conhecer outras oportunidades disponíveis em nossa plataforma que podem ser do seu interesse.

Até logo,
Equipe do Ciclo Carioca`,
		inscricao.Name,
		curso.Titulo,
	)

	return EmailTemplate{
		Subject: subject,
		Body:    body,
		IsHTML:  false,
	}
}
