package alelo

import "github.com/caarlos0/alelogo"
import log "github.com/Sirupsen/logrus"

func AllDetails(cpf, pwd string) (result []alelogo.CardDetails, err error) {
	client, err := alelogo.New(cpf, pwd)
	if err != nil {
		log.WithFields(log.Fields{
			"cpf": cpf,
		}).Info("Not logged in, telling user to do that")
		return result, err
	}
	cards, err := client.Cards()
	if err != nil {
		log.WithFields(log.Fields{
			"cpf": cpf,
		}).Error(err.Error())
		return result, err
	}
	log.WithFields(log.Fields{
		"cpf": cpf,
	}).Info("Got cards", cards)
	for _, card := range cards {
		details, err := client.Details(card)
		if err != nil {
			log.WithFields(log.Fields{
				"cpf":     cpf,
				"card_id": card.ID,
			}).Error(err.Error())
			return result, err
		}
		log.WithFields(log.Fields{
			"cpf":     cpf,
			"card_id": card.ID,
		}).Info("Got card details", details)
		result = append(result, details)
	}
	return result, err
}
